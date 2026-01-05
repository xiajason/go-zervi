package blockchain

// VersionStatusChangeRequest 版本状态变化请求
type VersionStatusChangeRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	VersionSource string `json:"version_source" binding:"required"`
	OldStatus     string `json:"old_status"`
	NewStatus     string `json:"new_status" binding:"required"`
	ChangeReason  string `json:"change_reason"`
	OperatorID    string `json:"operator_id"`
	Remark        string `json:"remark"`
}

// PermissionChangeRequest 权限变更请求
type PermissionChangeRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	VersionSource string `json:"version_source" binding:"required"`
	OldPermission string `json:"old_permission"`
	NewPermission string `json:"new_permission" binding:"required"`
	ChangeReason  string `json:"change_reason"`
	OperatorID    string `json:"operator_id"`
	Remark        string `json:"remark"`
}

// BlockchainResponse 区块链响应
type BlockchainResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ConsistencyValidationRequest 数据一致性校验请求
type ConsistencyValidationRequest struct {
	UserID        string `json:"user_id"`
	VersionSource string `json:"version_source"`
	CheckType     string `json:"check_type"` // FULL, INCREMENTAL, SPECIFIC
}

// TransactionQueryRequest 交易查询请求
type TransactionQueryRequest struct {
	Page            int    `json:"page"`
	Size            int    `json:"size"`
	TransactionType string `json:"transaction_type"`
	VersionSource   string `json:"version_source"`
	UserID          string `json:"user_id"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
}

// CrossVersionSyncRequest 跨版本同步请求
type CrossVersionSyncRequest struct {
	UserID         string   `json:"user_id" binding:"required"`
	SourceVersion  string   `json:"source_version" binding:"required"`
	TargetVersions []string `json:"target_versions" binding:"required"`
	SyncType       string   `json:"sync_type"` // STATUS, PERMISSION, FULL
	ForceSync      bool     `json:"force_sync"`
}

// CrossVersionSyncResponse 跨版本同步响应
type CrossVersionSyncResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// VersionStatusRecord 版本状态记录
type VersionStatusRecord struct {
	RecordID        string `json:"record_id"`
	UserID          string `json:"user_id"`
	VersionSource   string `json:"version_source"`
	OldStatus       string `json:"old_status"`
	NewStatus       string `json:"new_status"`
	ChangeReason    string `json:"change_reason"`
	OperatorID      string `json:"operator_id"`
	TransactionHash string `json:"transaction_hash"`
	BlockHeight     int64  `json:"block_height"`
	RecordTime      string `json:"record_time"`
}

// PermissionChangeRecord 权限变更记录
type PermissionChangeRecord struct {
	RecordID        string `json:"record_id"`
	UserID          string `json:"user_id"`
	VersionSource   string `json:"version_source"`
	OldPermission   string `json:"old_permission"`
	NewPermission   string `json:"new_permission"`
	ChangeReason    string `json:"change_reason"`
	OperatorID      string `json:"operator_id"`
	TransactionHash string `json:"transaction_hash"`
	BlockHeight     int64  `json:"block_height"`
	RecordTime      string `json:"record_time"`
}

// BlockchainTransaction 区块链交易
type BlockchainTransaction struct {
	TransactionID   string `json:"transaction_id"`
	TransactionHash string `json:"transaction_hash"`
	TransactionType string `json:"transaction_type"`
	VersionSource   string `json:"version_source"`
	UserID          string `json:"user_id"`
	OldStatus       string `json:"old_status"`
	NewStatus       string `json:"new_status"`
	ChangeReason    string `json:"change_reason"`
	OperatorID      string `json:"operator_id"`
	TransactionData string `json:"transaction_data"`
	Status          string `json:"status"`
	BlockHeight     int64  `json:"block_height"`
	CreateTime      string `json:"create_time"`
	ConfirmTime     string `json:"confirm_time"`
	Remark          string `json:"remark"`
}

// ConsistencyCheckResult 一致性校验结果
type ConsistencyCheckResult struct {
	ValidationTime      string                   `json:"validation_time"`
	Status              string                   `json:"status"`
	CheckedRecords      int                      `json:"checked_records"`
	Inconsistencies     int                      `json:"inconsistencies"`
	InconsistentRecords []map[string]interface{} `json:"inconsistent_records"`
	Message             string                   `json:"message"`
}

// VersionInfo 版本信息
type VersionInfo struct {
	VersionSource string `json:"version_source"`
	PortRange     string `json:"port_range"`
	Description   string `json:"description"`
	Status        string `json:"status"`
}

// CrossVersionUser 跨版本用户信息
type CrossVersionUser struct {
	UserID             string                 `json:"user_id"`
	Username           string                 `json:"username"`
	Email              string                 `json:"email"`
	VersionStatuses    map[string]string      `json:"version_statuses"`
	VersionPermissions map[string]string      `json:"version_permissions"`
	LastSyncTime       string                 `json:"last_sync_time"`
	Metadata           map[string]interface{} `json:"metadata"`
}
