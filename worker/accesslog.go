package worker

import (
	"bytes"
	"context"
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

	"github.com/szks-repo/usage-based-billing-sample/pkg/types"
)

// S3Uploader はログをS3にアップロードする責務を持つ
type S3Writer struct {
	s3Client   *s3.Client
	bucketName string

	logChan    chan types.ApiAccessLog
	buffer     []types.ApiAccessLog
	bufferSize int
	ticker     *time.Ticker
	mutex      sync.Mutex
	wg         sync.WaitGroup
	shutdown   chan struct{}
}

// NewS3Uploader は新しいUploaderインスタンスを作成
func NewS3Writer(
	client *s3.Client,
	bucket string,
	bufferSize int,
	interval time.Duration,
) *S3Writer {
	return &S3Writer{
		s3Client:   client,
		bucketName: bucket,
		logChan:    make(chan types.ApiAccessLog, bufferSize*2),
		buffer:     make([]types.ApiAccessLog, 0, bufferSize),
		bufferSize: bufferSize,
		ticker:     time.NewTicker(interval),
		shutdown:   make(chan struct{}),
	}
}

func (u *S3Writer) AddLog(log types.ApiAccessLog) {
	u.logChan <- log
}

func (u *S3Writer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				u.flush(context.WithoutCancel(ctx))
				slog.Info("ctx.Done", "error", ctx.Err())
				return
			case <-u.shutdown:
				u.flush(context.WithoutCancel(ctx))
				slog.Info("S3 uploader shutting down.")
				return
			case l := <-u.logChan:
				u.mutex.Lock()
				u.buffer = append(u.buffer, l)
				u.mutex.Unlock()

				if len(u.buffer) >= u.bufferSize {
					u.flush(ctx)
				}
			case <-u.ticker.C:
				u.flush(ctx)
			}
		}
	}()
}

func (u *S3Writer) Stop() {
	close(u.shutdown)
	u.wg.Wait()
}

func (u *S3Writer) flush(ctx context.Context) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if len(u.buffer) == 0 {
		return
	}

	logsToUpload := make([]types.ApiAccessLog, len(u.buffer))
	copy(logsToUpload, u.buffer)
	u.buffer = u.buffer[:0] // バッファをクリア

	slog.Info("Flushing logs to S3...", "numLogs", len(logsToUpload))

	// Parquetに変換
	parquetData, err := convertToParquet(logsToUpload)
	if err != nil {
		slog.Error("Error converting to parquet", "error", err)
		// 本番ではリトライ処理などを検討
		return
	}

	// S3にアップロード
	// logs/YYYY/MM/DD/uuid.parquet のようなキーにする
	key := fmt.Sprintf("logs/%s/%s.parquet", time.Now().Format("2006/01/02"), uuid.Must(uuid.NewV7()).String())

	if _, err = u.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &u.bucketName,
		Key:    &key,
		Body:   bytes.NewReader(parquetData),
	}); err != nil {
		slog.Error("Error uploading to S3", "error", err)
		return
	}

	slog.Info("Successfully uploaded", "key", key)
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

func convertToParquet(logs []types.ApiAccessLog) ([]byte, error) {
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
