package types

import "time"

type ApiAccessLog struct {
	AccountId  int64     `json:"account_id"`
	Timestamp  time.Time `json:"timestamp"`
	ClientIP   string    `json:"client_ip"`
	Path       string    `json:"path"`
	Method     string    `json:"method"`
	StatusCode int       `json:"status_code"`
	Latency    int64     `json:"latency"`
	UserAgent  string    `json:"user_agent"`
}
