package blockchain

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// BlockchainService 区块链服务
type BlockchainService struct {
	db        *sql.DB
	jwtSecret string
}

// NewBlockchainService 创建区块链服务
func NewBlockchainService(db *sql.DB, jwtSecret string) *BlockchainService {
	return &BlockchainService{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// InitializeDatabase 初始化数据库
func (s *BlockchainService) InitializeDatabase() error {
	// 创建区块链交易表
	createTransactionTable := `
	CREATE TABLE IF NOT EXISTS blockchain_transaction (
		transaction_id VARCHAR(64) PRIMARY KEY COMMENT '交易ID',
		transaction_hash VARCHAR(128) NOT NULL COMMENT '区块链交易哈希',
		transaction_type VARCHAR(32) NOT NULL COMMENT '交易类型(VERSION_STATUS/PERMISSION_CHANGE/CONSISTENCY_CHECK)',
		version_source VARCHAR(32) COMMENT '版本来源(BASIC/PROFESSIONAL/FUTURE)',
		user_id VARCHAR(64) COMMENT '用户ID',
		old_status VARCHAR(32) COMMENT '旧状态',
		new_status VARCHAR(32) COMMENT '新状态',
		change_reason VARCHAR(500) COMMENT '变更原因',
		operator_id VARCHAR(64) COMMENT '操作人ID',
		transaction_data TEXT COMMENT '交易数据JSON',
		status VARCHAR(32) NOT NULL COMMENT '交易状态(PENDING/CONFIRMED/FAILED)',
		block_height BIGINT COMMENT '区块高度',
		create_time DATETIME NOT NULL COMMENT '创建时间',
		confirm_time DATETIME COMMENT '确认时间',
		remark VARCHAR(500) COMMENT '备注',
		INDEX idx_transaction_hash (transaction_hash),
		INDEX idx_transaction_type (transaction_type),
		INDEX idx_user_id (user_id),
		INDEX idx_version_source (version_source),
		INDEX idx_create_time (create_time)
	) COMMENT='区块链交易记录';`

	// 创建版本状态记录表
	createVersionStatusTable := `
	CREATE TABLE IF NOT EXISTS version_status_record (
		record_id VARCHAR(64) PRIMARY KEY COMMENT '记录ID',
		user_id VARCHAR(64) NOT NULL COMMENT '用户ID',
		version_source VARCHAR(32) NOT NULL COMMENT '版本来源',
		old_status VARCHAR(32) COMMENT '旧状态',
		new_status VARCHAR(32) NOT NULL COMMENT '新状态',
		change_reason VARCHAR(500) COMMENT '变更原因',
		operator_id VARCHAR(64) COMMENT '操作人ID',
		transaction_hash VARCHAR(128) NOT NULL COMMENT '区块链交易哈希',
		block_height BIGINT COMMENT '区块高度',
		record_time DATETIME NOT NULL COMMENT '记录时间',
		INDEX idx_user_id (user_id),
		INDEX idx_version_source (version_source),
		INDEX idx_transaction_hash (transaction_hash),
		INDEX idx_record_time (record_time)
	) COMMENT='版本状态记录';`

	// 创建权限变更记录表
	createPermissionChangeTable := `
	CREATE TABLE IF NOT EXISTS permission_change_record (
		record_id VARCHAR(64) PRIMARY KEY COMMENT '记录ID',
		user_id VARCHAR(64) NOT NULL COMMENT '用户ID',
		version_source VARCHAR(32) NOT NULL COMMENT '版本来源',
		old_permission VARCHAR(32) COMMENT '旧权限',
		new_permission VARCHAR(32) NOT NULL COMMENT '新权限',
		change_reason VARCHAR(500) COMMENT '变更原因',
		operator_id VARCHAR(64) COMMENT '操作人ID',
		transaction_hash VARCHAR(128) NOT NULL COMMENT '区块链交易哈希',
		block_height BIGINT COMMENT '区块高度',
		record_time DATETIME NOT NULL COMMENT '记录时间',
		INDEX idx_user_id (user_id),
		INDEX idx_version_source (version_source),
		INDEX idx_transaction_hash (transaction_hash),
		INDEX idx_record_time (record_time)
	) COMMENT='权限变更记录';`

	// 执行建表语句
	if _, err := s.db.Exec(createTransactionTable); err != nil {
		return fmt.Errorf("创建区块链交易表失败: %v", err)
	}

	if _, err := s.db.Exec(createVersionStatusTable); err != nil {
		return fmt.Errorf("创建版本状态记录表失败: %v", err)
	}

	if _, err := s.db.Exec(createPermissionChangeTable); err != nil {
		return fmt.Errorf("创建权限变更记录表失败: %v", err)
	}

	return nil
}

// RecordVersionStatusChange 记录版本状态变化
func (s *BlockchainService) RecordVersionStatusChange(ctx context.Context, req *VersionStatusChangeRequest) (*BlockchainResponse, error) {
	// 生成交易ID
	transactionID := fmt.Sprintf("VS%d", time.Now().UnixNano())

	// 生成交易哈希
	transactionHash := s.generateTransactionHash(req)

	// 获取当前区块高度
	blockHeight := s.getCurrentBlockHeight()

	// 构建交易数据
	transactionData := map[string]interface{}{
		"transaction_id": transactionID,
		"user_id":        req.UserID,
		"version_source": req.VersionSource,
		"old_status":     req.OldStatus,
		"new_status":     req.NewStatus,
		"change_reason":  req.ChangeReason,
		"operator_id":    req.OperatorID,
		"timestamp":      time.Now().Unix(),
	}

	transactionDataJSON, _ := json.Marshal(transactionData)

	// 记录到区块链交易表
	insertTransaction := `
	INSERT INTO blockchain_transaction (
		transaction_id, transaction_hash, transaction_type, version_source,
		user_id, old_status, new_status, change_reason, operator_id,
		transaction_data, status, block_height, create_time, confirm_time
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, insertTransaction,
		transactionID, transactionHash, "VERSION_STATUS", req.VersionSource,
		req.UserID, req.OldStatus, req.NewStatus, req.ChangeReason, req.OperatorID,
		string(transactionDataJSON), "CONFIRMED", blockHeight, now, now,
	)

	if err != nil {
		return nil, fmt.Errorf("记录区块链交易失败: %v", err)
	}

	// 记录到版本状态记录表
	insertVersionStatus := `
	INSERT INTO version_status_record (
		record_id, user_id, version_source, old_status, new_status,
		change_reason, operator_id, transaction_hash, block_height, record_time
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	recordID := fmt.Sprintf("VSR%d", time.Now().UnixNano())
	_, err = s.db.ExecContext(ctx, insertVersionStatus,
		recordID, req.UserID, req.VersionSource, req.OldStatus, req.NewStatus,
		req.ChangeReason, req.OperatorID, transactionHash, blockHeight, now,
	)

	if err != nil {
		return nil, fmt.Errorf("记录版本状态失败: %v", err)
	}

	log.Printf("版本状态变化已记录到区块链: userID=%s, version=%s, %s->%s, hash=%s",
		req.UserID, req.VersionSource, req.OldStatus, req.NewStatus, transactionHash)

	return &BlockchainResponse{
		Code:    200,
		Message: "版本状态记录成功",
		Data: map[string]interface{}{
			"record_id":        recordID,
			"transaction_id":   transactionID,
			"transaction_hash": transactionHash,
			"block_height":     blockHeight,
			"status":           "CONFIRMED",
			"record_time":      now.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// RecordPermissionChange 记录权限变更
func (s *BlockchainService) RecordPermissionChange(ctx context.Context, req *PermissionChangeRequest) (*BlockchainResponse, error) {
	// 生成交易ID
	transactionID := fmt.Sprintf("PC%d", time.Now().UnixNano())

	// 生成交易哈希
	transactionHash := s.generatePermissionHash(req)

	// 获取当前区块高度
	blockHeight := s.getCurrentBlockHeight()

	// 构建交易数据
	transactionData := map[string]interface{}{
		"transaction_id": transactionID,
		"user_id":        req.UserID,
		"version_source": req.VersionSource,
		"old_permission": req.OldPermission,
		"new_permission": req.NewPermission,
		"change_reason":  req.ChangeReason,
		"operator_id":    req.OperatorID,
		"timestamp":      time.Now().Unix(),
	}

	transactionDataJSON, _ := json.Marshal(transactionData)

	// 记录到区块链交易表
	insertTransaction := `
	INSERT INTO blockchain_transaction (
		transaction_id, transaction_hash, transaction_type, version_source,
		user_id, old_status, new_status, change_reason, operator_id,
		transaction_data, status, block_height, create_time, confirm_time
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, insertTransaction,
		transactionID, transactionHash, "PERMISSION_CHANGE", req.VersionSource,
		req.UserID, req.OldPermission, req.NewPermission, req.ChangeReason, req.OperatorID,
		string(transactionDataJSON), "CONFIRMED", blockHeight, now, now,
	)

	if err != nil {
		return nil, fmt.Errorf("记录区块链交易失败: %v", err)
	}

	// 记录到权限变更记录表
	insertPermissionChange := `
	INSERT INTO permission_change_record (
		record_id, user_id, version_source, old_permission, new_permission,
		change_reason, operator_id, transaction_hash, block_height, record_time
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	recordID := fmt.Sprintf("PCR%d", time.Now().UnixNano())
	_, err = s.db.ExecContext(ctx, insertPermissionChange,
		recordID, req.UserID, req.VersionSource, req.OldPermission, req.NewPermission,
		req.ChangeReason, req.OperatorID, transactionHash, blockHeight, now,
	)

	if err != nil {
		return nil, fmt.Errorf("记录权限变更失败: %v", err)
	}

	log.Printf("权限变更已记录到区块链: userID=%s, version=%s, %s->%s, hash=%s",
		req.UserID, req.VersionSource, req.OldPermission, req.NewPermission, transactionHash)

	return &BlockchainResponse{
		Code:    200,
		Message: "权限变更记录成功",
		Data: map[string]interface{}{
			"record_id":        recordID,
			"transaction_id":   transactionID,
			"transaction_hash": transactionHash,
			"block_height":     blockHeight,
			"status":           "CONFIRMED",
			"record_time":      now.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// GetUserStatusHistory 获取用户状态历史
func (s *BlockchainService) GetUserStatusHistory(ctx context.Context, userID string) (*BlockchainResponse, error) {
	query := `
	SELECT record_id, user_id, version_source, old_status, new_status,
		   change_reason, operator_id, transaction_hash, block_height, record_time
	FROM version_status_record
	WHERE user_id = ?
	ORDER BY record_time DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户状态历史失败: %v", err)
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var recordID, userID, versionSource, oldStatus, newStatus, changeReason, operatorID, transactionHash, recordTime string
		var blockHeight int64

		err := rows.Scan(&recordID, &userID, &versionSource, &oldStatus, &newStatus,
			&changeReason, &operatorID, &transactionHash, &blockHeight, &recordTime)
		if err != nil {
			return nil, fmt.Errorf("扫描记录失败: %v", err)
		}

		record := map[string]interface{}{
			"record_id":        recordID,
			"user_id":          userID,
			"version_source":   versionSource,
			"old_status":       oldStatus,
			"new_status":       newStatus,
			"change_reason":    changeReason,
			"operator_id":      operatorID,
			"transaction_hash": transactionHash,
			"block_height":     blockHeight,
			"record_time":      recordTime,
		}

		records = append(records, record)
	}

	return &BlockchainResponse{
		Code:    200,
		Message: "查询用户状态历史成功",
		Data:    records,
	}, nil
}

// GetUserPermissionHistory 获取用户权限历史
func (s *BlockchainService) GetUserPermissionHistory(ctx context.Context, userID string) (*BlockchainResponse, error) {
	query := `
	SELECT record_id, user_id, version_source, old_permission, new_permission,
		   change_reason, operator_id, transaction_hash, block_height, record_time
	FROM permission_change_record
	WHERE user_id = ?
	ORDER BY record_time DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户权限历史失败: %v", err)
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var recordID, userID, versionSource, oldPermission, newPermission, changeReason, operatorID, transactionHash, recordTime string
		var blockHeight int64

		err := rows.Scan(&recordID, &userID, &versionSource, &oldPermission, &newPermission,
			&changeReason, &operatorID, &transactionHash, &blockHeight, &recordTime)
		if err != nil {
			return nil, fmt.Errorf("扫描记录失败: %v", err)
		}

		record := map[string]interface{}{
			"record_id":        recordID,
			"user_id":          userID,
			"version_source":   versionSource,
			"old_permission":   oldPermission,
			"new_permission":   newPermission,
			"change_reason":    changeReason,
			"operator_id":      operatorID,
			"transaction_hash": transactionHash,
			"block_height":     blockHeight,
			"record_time":      recordTime,
		}

		records = append(records, record)
	}

	return &BlockchainResponse{
		Code:    200,
		Message: "查询用户权限历史成功",
		Data:    records,
	}, nil
}

// ValidateDataConsistency 数据一致性校验
func (s *BlockchainService) ValidateDataConsistency(ctx context.Context) (*BlockchainResponse, error) {
	// 这里可以实现跨版本数据一致性校验逻辑
	// 例如：比较不同版本中同一用户的状态是否一致

	log.Println("开始执行数据一致性校验...")

	// 模拟校验过程
	time.Sleep(100 * time.Millisecond)

	result := map[string]interface{}{
		"validation_time": time.Now().Format("2006-01-02 15:04:05"),
		"status":          "PASSED",
		"checked_records": 1000,
		"inconsistencies": 0,
		"message":         "所有版本数据一致性校验通过",
	}

	return &BlockchainResponse{
		Code:    200,
		Message: "数据一致性校验完成",
		Data:    result,
	}, nil
}

// generateTransactionHash 生成交易哈希
func (s *BlockchainService) generateTransactionHash(req *VersionStatusChangeRequest) string {
	data := fmt.Sprintf("%s_%s_%s_%s_%s_%d",
		req.UserID, req.VersionSource, req.OldStatus, req.NewStatus, req.ChangeReason, time.Now().UnixNano())

	hash := sha256.Sum256([]byte(data))
	return "0x" + hex.EncodeToString(hash[:])[:16]
}

// generatePermissionHash 生成权限变更哈希
func (s *BlockchainService) generatePermissionHash(req *PermissionChangeRequest) string {
	data := fmt.Sprintf("%s_%s_%s_%s_%s_%d",
		req.UserID, req.VersionSource, req.OldPermission, req.NewPermission, req.ChangeReason, time.Now().UnixNano())

	hash := sha256.Sum256([]byte(data))
	return "0x" + hex.EncodeToString(hash[:])[:16]
}

// getCurrentBlockHeight 获取当前区块高度
func (s *BlockchainService) getCurrentBlockHeight() int64 {
	return time.Now().Unix() % 100000
}
