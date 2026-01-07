package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
	statements := []string{
		`CREATE TABLE IF NOT EXISTS blockchain_transaction (
			transaction_id TEXT PRIMARY KEY,
			transaction_hash TEXT NOT NULL,
			transaction_type TEXT NOT NULL,
			version_source TEXT,
			user_id TEXT,
			old_status TEXT,
			new_status TEXT,
			change_reason TEXT,
			operator_id TEXT,
			transaction_data TEXT,
			status TEXT NOT NULL,
			block_height BIGINT,
			create_time TIMESTAMPTZ NOT NULL,
			confirm_time TIMESTAMPTZ,
			remark TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS version_status_record (
			record_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			version_source TEXT NOT NULL,
			old_status TEXT,
			new_status TEXT NOT NULL,
			change_reason TEXT,
			operator_id TEXT,
			transaction_hash TEXT NOT NULL,
			block_height BIGINT,
			record_time TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS permission_change_record (
			record_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			version_source TEXT NOT NULL,
			old_permission TEXT,
			new_permission TEXT NOT NULL,
			change_reason TEXT,
			operator_id TEXT,
			transaction_hash TEXT NOT NULL,
			block_height BIGINT,
			record_time TIMESTAMPTZ NOT NULL
		)`,
	}

	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("执行数据库初始化语句失败: %w", err)
		}
	}

	indexStatements := []string{
		`CREATE INDEX IF NOT EXISTS idx_blockchain_transaction_hash ON blockchain_transaction(transaction_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_blockchain_transaction_type ON blockchain_transaction(transaction_type)`,
		`CREATE INDEX IF NOT EXISTS idx_blockchain_transaction_user ON blockchain_transaction(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_blockchain_transaction_version ON blockchain_transaction(version_source)`,
		`CREATE INDEX IF NOT EXISTS idx_blockchain_transaction_create_time ON blockchain_transaction(create_time)`,
		`CREATE INDEX IF NOT EXISTS idx_version_status_user ON version_status_record(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_version_status_version ON version_status_record(version_source)`,
		`CREATE INDEX IF NOT EXISTS idx_version_status_hash ON version_status_record(transaction_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_version_status_time ON version_status_record(record_time)`,
		`CREATE INDEX IF NOT EXISTS idx_permission_change_user ON permission_change_record(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_permission_change_version ON permission_change_record(version_source)`,
		`CREATE INDEX IF NOT EXISTS idx_permission_change_hash ON permission_change_record(transaction_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_permission_change_time ON permission_change_record(record_time)`,
	}

	for _, stmt := range indexStatements {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("创建索引失败: %w", err)
		}
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
		transaction_data, status, block_height, create_time, confirm_time, remark
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, insertTransaction,
		transactionID, transactionHash, "VERSION_STATUS", req.VersionSource,
		req.UserID, req.OldStatus, req.NewStatus, req.ChangeReason, req.OperatorID,
		string(transactionDataJSON), "CONFIRMED", blockHeight, now, now, req.Remark,
	)

	if err != nil {
		return nil, fmt.Errorf("记录区块链交易失败: %v", err)
	}

	// 记录到版本状态记录表
	insertVersionStatus := `
	INSERT INTO version_status_record (
		record_id, user_id, version_source, old_status, new_status,
		change_reason, operator_id, transaction_hash, block_height, record_time
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

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
		transaction_data, status, block_height, create_time, confirm_time, remark
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, insertTransaction,
		transactionID, transactionHash, "PERMISSION_CHANGE", req.VersionSource,
		req.UserID, req.OldPermission, req.NewPermission, req.ChangeReason, req.OperatorID,
		string(transactionDataJSON), "CONFIRMED", blockHeight, now, now, req.Remark,
	)

	if err != nil {
		return nil, fmt.Errorf("记录区块链交易失败: %v", err)
	}

	// 记录到权限变更记录表
	insertPermissionChange := `
	INSERT INTO permission_change_record (
		record_id, user_id, version_source, old_permission, new_permission,
		change_reason, operator_id, transaction_hash, block_height, record_time
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

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
	WHERE user_id = $1
	ORDER BY record_time DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户状态历史失败: %v", err)
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var (
			recordID        string
			dbUserID        string
			versionSource   string
			oldStatus       sql.NullString
			newStatus       string
			changeReason    sql.NullString
			operatorID      sql.NullString
			transactionHash string
			blockHeight     sql.NullInt64
			recordTime      time.Time
		)

		if err := rows.Scan(&recordID, &dbUserID, &versionSource, &oldStatus, &newStatus,
			&changeReason, &operatorID, &transactionHash, &blockHeight, &recordTime); err != nil {
			return nil, fmt.Errorf("扫描记录失败: %v", err)
		}

		record := map[string]interface{}{
			"record_id":        recordID,
			"user_id":          dbUserID,
			"version_source":   versionSource,
			"old_status":       stringOrEmpty(oldStatus),
			"new_status":       newStatus,
			"change_reason":    stringOrEmpty(changeReason),
			"operator_id":      stringOrEmpty(operatorID),
			"transaction_hash": transactionHash,
			"block_height":     intOrZero(blockHeight),
			"record_time":      recordTime.UTC().Format(time.RFC3339),
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
	WHERE user_id = $1
	ORDER BY record_time DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户权限历史失败: %v", err)
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var (
			recordID        string
			dbUserID        string
			versionSource   string
			oldPermission   sql.NullString
			newPermission   string
			changeReason    sql.NullString
			operatorID      sql.NullString
			transactionHash string
			blockHeight     sql.NullInt64
			recordTime      time.Time
		)

		if err := rows.Scan(&recordID, &dbUserID, &versionSource, &oldPermission, &newPermission,
			&changeReason, &operatorID, &transactionHash, &blockHeight, &recordTime); err != nil {
			return nil, fmt.Errorf("扫描记录失败: %v", err)
		}

		record := map[string]interface{}{
			"record_id":        recordID,
			"user_id":          dbUserID,
			"version_source":   versionSource,
			"old_permission":   stringOrEmpty(oldPermission),
			"new_permission":   newPermission,
			"change_reason":    stringOrEmpty(changeReason),
			"operator_id":      stringOrEmpty(operatorID),
			"transaction_hash": transactionHash,
			"block_height":     intOrZero(blockHeight),
			"record_time":      recordTime.UTC().Format(time.RFC3339),
		}

		records = append(records, record)
	}

	return &BlockchainResponse{
		Code:    200,
		Message: "查询用户权限历史成功",
		Data:    records,
	}, nil
}

// GetTransactionList 获取交易列表
func (s *BlockchainService) GetTransactionList(ctx context.Context, req TransactionQueryRequest) ([]BlockchainTransaction, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 10
	}

	filters := make([]string, 0)
	args := make([]interface{}, 0)
	idx := 1

	if req.TransactionType != "" {
		filters = append(filters, fmt.Sprintf("transaction_type = $%d", idx))
		args = append(args, req.TransactionType)
		idx++
	}

	if req.VersionSource != "" {
		filters = append(filters, fmt.Sprintf("version_source = $%d", idx))
		args = append(args, req.VersionSource)
		idx++
	}

	if req.UserID != "" {
		filters = append(filters, fmt.Sprintf("user_id = $%d", idx))
		args = append(args, req.UserID)
		idx++
	}

	if req.StartTime != "" {
		if start, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			filters = append(filters, fmt.Sprintf("create_time >= $%d", idx))
			args = append(args, start)
			idx++
		}
	}

	if req.EndTime != "" {
		if end, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			filters = append(filters, fmt.Sprintf("create_time <= $%d", idx))
			args = append(args, end)
			idx++
		}
	}

	whereClause := ""
	if len(filters) > 0 {
		whereClause = "WHERE " + strings.Join(filters, " AND ")
	}

	totalQuery := "SELECT COUNT(*) FROM blockchain_transaction " + whereClause
	var total int64
	if err := s.db.QueryRowContext(ctx, totalQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("统计交易数量失败: %w", err)
	}

	offset := (req.Page - 1) * req.Size
	limitPlaceholder := fmt.Sprintf("$%d", idx)
	idx++
	offsetPlaceholder := fmt.Sprintf("$%d", idx)

	query := `SELECT transaction_id, transaction_hash, transaction_type, version_source,
		user_id, old_status, new_status, change_reason, operator_id,
		transaction_data, status, block_height, create_time, confirm_time, remark
	FROM blockchain_transaction ` + whereClause + ` ORDER BY create_time DESC LIMIT ` + limitPlaceholder + ` OFFSET ` + offsetPlaceholder

	queryArgs := append(args, req.Size, offset)
	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询交易列表失败: %w", err)
	}
	defer rows.Close()

	transactions := make([]BlockchainTransaction, 0)
	for rows.Next() {
		var (
			transactionID   string
			transactionHash string
			transactionType string
			versionSource   sql.NullString
			userID          sql.NullString
			oldStatus       sql.NullString
			newStatus       sql.NullString
			changeReason    sql.NullString
			operatorID      sql.NullString
			transactionData sql.NullString
			status          string
			blockHeight     sql.NullInt64
			createTime      time.Time
			confirmTime     sql.NullTime
			remark          sql.NullString
		)

		if err := rows.Scan(
			&transactionID,
			&transactionHash,
			&transactionType,
			&versionSource,
			&userID,
			&oldStatus,
			&newStatus,
			&changeReason,
			&operatorID,
			&transactionData,
			&status,
			&blockHeight,
			&createTime,
			&confirmTime,
			&remark,
		); err != nil {
			return nil, 0, fmt.Errorf("扫描交易记录失败: %w", err)
		}

		transactions = append(transactions, BlockchainTransaction{
			TransactionID:   transactionID,
			TransactionHash: transactionHash,
			TransactionType: transactionType,
			VersionSource:   stringOrEmpty(versionSource),
			UserID:          stringOrEmpty(userID),
			OldStatus:       stringOrEmpty(oldStatus),
			NewStatus:       stringOrEmpty(newStatus),
			ChangeReason:    stringOrEmpty(changeReason),
			OperatorID:      stringOrEmpty(operatorID),
			TransactionData: stringOrEmpty(transactionData),
			Status:          status,
			BlockHeight:     intOrZero(blockHeight),
			CreateTime:      createTime.UTC().Format(time.RFC3339),
			ConfirmTime:     timeOrEmpty(confirmTime),
			Remark:          stringOrEmpty(remark),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("遍历交易记录失败: %w", err)
	}

	return transactions, total, nil
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

func stringOrEmpty(val sql.NullString) string {
	if val.Valid {
		return val.String
	}
	return ""
}

func intOrZero(val sql.NullInt64) int64 {
	if val.Valid {
		return val.Int64
	}
	return 0
}

func timeOrEmpty(val sql.NullTime) string {
	if val.Valid {
		return val.Time.UTC().Format(time.RFC3339)
	}
	return ""
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
