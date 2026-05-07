// Application scenarios:
// - Hold the shared transaction models used by distributed transaction implementations.
// - Standardize transaction inspection, branch configuration, and SAGA step descriptions.
// - Keep transaction-related DTOs reusable across providers and tooling.
//
// 适用场景：
// - 承载分布式事务实现共享的事务模型。
// - 统一事务查询、分支配置和 SAGA 步骤描述。
// - 让事务相关 DTO 可在 provider 和工具链之间复用。
package integration

// TransactionInfo describes one distributed transaction snapshot.
//
// TransactionInfo 描述一份分布式事务快照。
type TransactionInfo struct {
	GID             string
	Status          string
	TransactionType string
	CreateTime      int64
	UpdateTime      int64
	Steps           []TransactionStep
}

// TransactionStep describes one branch or step in a transaction.
//
// TransactionStep 描述事务中的一个分支或步骤。
type TransactionStep struct {
	BranchID string
	Status   string
	Op       string
	URL      string
}

// SAGATransaction describes a built SAGA transaction.
//
// SAGATransaction 描述一份构建好的 SAGA 事务。
type SAGATransaction struct {
	GID      string
	Steps    []SAGAStep
	Payloads []any
}

// SAGAStep describes one SAGA branch step.
//
// SAGAStep 描述一个 SAGA 分支步骤。
type SAGAStep struct {
	Action        string
	Compensate    string
	Payload       any
	RetryCount    int
	RetryInterval int
	Timeout       int
}

// BranchOptions describes branch-level retry and timeout settings.
//
// BranchOptions 描述分支级重试和超时设置。
type BranchOptions struct {
	RetryCount    int
	RetryInterval int
	Timeout       int
}
