package integration

type TransactionInfo struct {
	GID             string
	Status          string
	TransactionType string
	CreateTime      int64
	UpdateTime      int64
	Steps           []TransactionStep
}

type TransactionStep struct {
	BranchID string
	Status   string
	Op       string
	URL      string
}

type SAGATransaction struct {
	GID      string
	Steps    []SAGAStep
	Payloads []any
}

type SAGAStep struct {
	Action        string
	Compensate    string
	Payload       any
	RetryCount    int
	RetryInterval int
	Timeout       int
}

type BranchOptions struct {
	RetryCount    int
	RetryInterval int
	Timeout       int
}
