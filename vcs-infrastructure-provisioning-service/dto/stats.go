package dto

type PostgreSQLStatsResponse struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsed    int64   `json:"memory_used"`
	MemoryLimit   int64   `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     int64   `json:"network_rx"`
	NetworkTx     int64   `json:"network_tx"`
	DiskRead      int64   `json:"disk_read"`
	DiskWrite     int64   `json:"disk_write"`
}
