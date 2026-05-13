package response

type Order struct {
	ID          uint    `json:"id"`
	UserID      uint    `json:"user_id"`
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	TotalPrice  float64 `json:"total_price"`
	Status      string  `json:"status"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}

type OrderList struct {
	Items []Order `json:"items"`
	Total int64   `json:"total"`
	Page  int     `json:"page"`
	Size  int     `json:"size"`
}

type RemoteUser struct {
	ID                uint64 `json:"id"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	TraceID           string `json:"trace_id"`
	RequestID         string `json:"request_id"`
	MetadataDemo      string `json:"metadata_demo"`
	CallerServiceName string `json:"caller_service_name"`
}

type LockDemo struct {
	Key     string `json:"key"`
	Success bool   `json:"success"`
}
