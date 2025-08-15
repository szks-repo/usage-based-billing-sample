package types

import "time"

type AccessLog struct {
	Timestamp  time.Time `json:"timestamp" parquet:"name=timestamp, type=TIMESTAMP_MILLIS"`
	ClientIP   string    `json:"client_ip" parquet:"name=client_ip, type=UTF8"`
	Path       string    `json:"path" parquet:"name=path, type=UTF8"`
	Method     string    `json:"method" parquet:"name=method, type=UTF8"`
	Protocol   string    `json:"protocol" parquet:"name=protocol, type=UTF8"`
	StatusCode int       `json:"status_code" parquet:"name=status_code, type=INT32"`
	Latency    int64     `json:"latency" parquet:"name=latency_ms, type=INT64"`
	UserAgent  string    `json:"user_agent" parquet:"name=user_agent, type=UTF8"`
}
