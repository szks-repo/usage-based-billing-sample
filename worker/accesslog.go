package worker

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
	"github.com/apache/arrow/go/v17/parquet"
	"github.com/apache/arrow/go/v17/parquet/compress"
	"github.com/apache/arrow/go/v17/parquet/pqarrow"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/szks-repo/usage-based-billing-sample/pkg/db"
	"github.com/szks-repo/usage-based-billing-sample/pkg/db/dto"
	"github.com/szks-repo/usage-based-billing-sample/pkg/types"
)

type AccessLogRecorder struct {
	s3Client   *s3.Client
	bucketName string

	dbConn *sql.DB

	logChan    chan types.ApiAccessLog
	buffer     []types.ApiAccessLog
	bufferSize int
	ticker     *time.Ticker
	mutex      sync.Mutex
	wg         sync.WaitGroup
	shutdown   chan struct{}
}

// NewS3Uploader は新しいUploaderインスタンスを作成
func NewAccessLogRecorder(
	client *s3.Client,
	bucket string,
	bufferSize int,
	interval time.Duration,
	dbConn *sql.DB,
) *AccessLogRecorder {
	return &AccessLogRecorder{
		s3Client:   client,
		bucketName: bucket,
		dbConn:     dbConn,
		logChan:    make(chan types.ApiAccessLog, bufferSize*2),
		buffer:     make([]types.ApiAccessLog, 0, bufferSize),
		bufferSize: bufferSize,
		ticker:     time.NewTicker(interval),
		shutdown:   make(chan struct{}),
	}
}

func (u *AccessLogRecorder) Push(log types.ApiAccessLog) {
	u.logChan <- log
}

func (r *AccessLogRecorder) Observe(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				r.flush(context.WithoutCancel(ctx))
				slog.Info("ctx.Done", "error", ctx.Err())
				return
			case <-r.shutdown:
				r.flush(context.WithoutCancel(ctx))
				slog.Info("S3 uploader shutting down.")
				return
			case l := <-r.logChan:
				r.mutex.Lock()
				r.buffer = append(r.buffer, l)
				r.mutex.Unlock()

				if len(r.buffer) >= r.bufferSize {
					r.flush(ctx)
				}
			case <-r.ticker.C:
				r.flush(ctx)
			}
		}
	}()
}

func (r *AccessLogRecorder) Stop() {
	close(r.shutdown)
	r.wg.Wait()
}

func (r *AccessLogRecorder) uploadToS3(ctx context.Context, logs []types.ApiAccessLog) {
	slog.Info("Flushing logs to S3...", "numLogs", len(logs))

	// Parquetに変換
	parquetData, err := r.convertToParquet(logs)
	if err != nil {
		slog.Error("Error converting to parquet", "error", err)
		// 本番ではリトライ処理などを検討
		return
	}

	// S3にアップロード
	// logs/YYYY/MM/DD/uuid.parquet のようなキーにする
	key := fmt.Sprintf("logs/%s/%s.parquet", time.Now().Format("2006/01/02"), uuid.Must(uuid.NewV7()).String())

	if _, err = r.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &r.bucketName,
		Key:    &key,
		Body:   bytes.NewReader(parquetData),
	}); err != nil {
		slog.Error("Error uploading to S3", "error", err)
		return
	}

	slog.Info("Successfully uploaded", "key", key)
}

func (r *AccessLogRecorder) flush(ctx context.Context) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.buffer) == 0 {
		return
	}

	logsToUpload := make([]types.ApiAccessLog, len(r.buffer))
	copy(logsToUpload, r.buffer)
	r.buffer = r.buffer[:0]

	var wg sync.WaitGroup
	wg.Go(func() {
		r.uploadToS3(ctx, logsToUpload)
	})
	wg.Go(func() {
		r.saveAggregated(ctx, logsToUpload)
	})
	wg.Wait()
}

var parquetSchema = arrow.NewSchema(
	[]arrow.Field{
		{Name: "account_id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "client_ip", Type: arrow.BinaryTypes.String},
		{Name: "method", Type: arrow.BinaryTypes.String},
		{Name: "path", Type: arrow.BinaryTypes.String},
		{Name: "status_code", Type: arrow.PrimitiveTypes.Int32},
		{Name: "latency_ms", Type: arrow.PrimitiveTypes.Int64},
		{Name: "user_agent", Type: arrow.BinaryTypes.String},
		{Name: "timestamp", Type: arrow.FixedWidthTypes.Timestamp_ms},
	},
	nil, // metadata
)

func (a *AccessLogRecorder) convertToParquet(logs []types.ApiAccessLog) ([]byte, error) {
	pool := memory.NewGoAllocator()
	rb := array.NewRecordBuilder(pool, parquetSchema)
	defer rb.Release()

	for _, l := range logs {
		rb.Field(0).(*array.Int64Builder).Append(l.AccountId)
		rb.Field(1).(*array.StringBuilder).Append(l.ClientIP)
		rb.Field(2).(*array.StringBuilder).Append(l.Method)
		rb.Field(3).(*array.StringBuilder).Append(l.Path)
		rb.Field(4).(*array.Int32Builder).Append(int32(l.StatusCode))
		rb.Field(5).(*array.Int64Builder).Append(l.Latency)
		rb.Field(6).(*array.StringBuilder).Append(l.UserAgent)
		rb.Field(7).(*array.TimestampBuilder).Append(arrow.Timestamp(l.Timestamp.UnixMilli()))
	}

	rec := rb.NewRecord()
	defer rec.Release()

	var buf bytes.Buffer
	props := parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Snappy))
	writer, err := pqarrow.NewFileWriter(parquetSchema, &buf, props, pqarrow.NewArrowWriterProperties())
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet file writer: %w", err)
	}

	if err := writer.Write(rec); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to write record to parquet: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (r *AccessLogRecorder) saveAggregated(ctx context.Context, accessLogs []types.ApiAccessLog) error {
	if len(accessLogs) == 0 {
		return nil
	}

	groupByAccounts := make(map[int64][]types.ApiAccessLog)
	for _, l := range accessLogs {
		groupByAccounts[l.AccountId] = append(groupByAccounts[l.AccountId], l)
	}

	var dst []*dto.EveryMinuteAPIUsage
	for accountId, logs := range groupByAccounts {
		minuteGroup := make(map[string]int)
		for _, l := range logs {
			minuteGroup[l.Timestamp.Format("200601021504")] += 1
		}
		for minute, usage := range minuteGroup {
			dst = append(dst, &dto.EveryMinuteAPIUsage{
				AccountID: uint64(accountId),
				Minute:    minute,
				Usage:     uint64(usage),
			})
		}
	}

	if len(dst) == 0 {
		return nil
	}

	slog.Info("Upsert minute aggregate records", "num", len(dst))

	args := make([]any, 0, len(dst)*3)
	for _, v := range dst {
		args = append(args, v.AccountID, v.Minute, v.Usage)
	}

	result, err := r.dbConn.ExecContext(
		ctx,
		"INSERT INTO every_minute_api_usage (`account_id`, `minute`, `usage`) "+db.MakeValues(3, len(dst))+" "+
			"ON DUPLICATE KEY UPDATE "+
			"`usage` = `usage` + VALUES(`usage`), `updated_at` = NOW()",
		args...,
	)
	if err != nil {
		slog.Error("Failed to ExecContext", "error", err)
		return err
	}
	ra, _ := result.RowsAffected()
	slog.Info("Upsert every_minute_api_usage", "rowsAffected", ra)

	return nil
}
