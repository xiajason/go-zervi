package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jobfirst/jobfirst-core/blockchain"
)

// 示例：如何在其他微服务中集成区块链服务

func main() {
	// 创建区块链服务客户端
	blockchainClient := blockchain.NewClient("http://localhost:8208")

	// 示例1：记录用户状态变化
	recordUserStatusChange(blockchainClient)

	// 示例2：记录权限变更
	recordPermissionChange(blockchainClient)

	// 示例3：查询用户历史
	queryUserHistory(blockchainClient)

	// 示例4：数据一致性校验
	validateConsistency(blockchainClient)
}

// recordUserStatusChange 记录用户状态变化示例
func recordUserStatusChange(client *blockchain.Client) {
	fmt.Println("=== 示例1：记录用户状态变化 ===")

	req := &blockchain.VersionStatusChangeRequest{
		UserID:        "user_001",
		VersionSource: "BASIC",
		OldStatus:     "ACTIVE",
		NewStatus:     "INACTIVE",
		ChangeReason:  "用户注销账户",
		OperatorID:    "admin_001",
		Remark:        "系统自动处理",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.RecordVersionStatusChange(ctx, req)
	if err != nil {
		log.Printf("记录用户状态变化失败: %v", err)
		return
	}

	fmt.Printf("✅ 用户状态变化记录成功:\n")
	fmt.Printf("   交易哈希: %v\n", resp.Data.(map[string]interface{})["transaction_hash"])
	fmt.Printf("   区块高度: %v\n", resp.Data.(map[string]interface{})["block_height"])
	fmt.Printf("   记录时间: %v\n", resp.Data.(map[string]interface{})["record_time"])
	fmt.Println()
}

// recordPermissionChange 记录权限变更示例
func recordPermissionChange(client *blockchain.Client) {
	fmt.Println("=== 示例2：记录权限变更 ===")

	req := &blockchain.PermissionChangeRequest{
		UserID:        "user_001",
		VersionSource: "PROFESSIONAL",
		OldPermission: "READ",
		NewPermission: "WRITE",
		ChangeReason:  "用户升级到专业版",
		OperatorID:    "admin_001",
		Remark:        "权限升级",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.RecordPermissionChange(ctx, req)
	if err != nil {
		log.Printf("记录权限变更失败: %v", err)
		return
	}

	fmt.Printf("✅ 权限变更记录成功:\n")
	fmt.Printf("   交易哈希: %v\n", resp.Data.(map[string]interface{})["transaction_hash"])
	fmt.Printf("   区块高度: %v\n", resp.Data.(map[string]interface{})["block_height"])
	fmt.Printf("   记录时间: %v\n", resp.Data.(map[string]interface{})["record_time"])
	fmt.Println()
}

// queryUserHistory 查询用户历史示例
func queryUserHistory(client *blockchain.Client) {
	fmt.Println("=== 示例3：查询用户历史 ===")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 查询用户状态历史
	statusResp, err := client.GetUserStatusHistory(ctx, "user_001")
	if err != nil {
		log.Printf("查询用户状态历史失败: %v", err)
		return
	}

	fmt.Printf("✅ 用户状态历史查询成功:\n")
	if records, ok := statusResp.Data.([]map[string]interface{}); ok {
		for i, record := range records {
			fmt.Printf("   记录 %d: %s -> %s (%s)\n",
				i+1,
				record["old_status"],
				record["new_status"],
				record["record_time"])
		}
	}

	// 查询用户权限历史
	permissionResp, err := client.GetUserPermissionHistory(ctx, "user_001")
	if err != nil {
		log.Printf("查询用户权限历史失败: %v", err)
		return
	}

	fmt.Printf("✅ 用户权限历史查询成功:\n")
	if records, ok := permissionResp.Data.([]map[string]interface{}); ok {
		for i, record := range records {
			fmt.Printf("   记录 %d: %s -> %s (%s)\n",
				i+1,
				record["old_permission"],
				record["new_permission"],
				record["record_time"])
		}
	}
	fmt.Println()
}

// validateConsistency 数据一致性校验示例
func validateConsistency(client *blockchain.Client) {
	fmt.Println("=== 示例4：数据一致性校验 ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.ValidateDataConsistency(ctx)
	if err != nil {
		log.Printf("数据一致性校验失败: %v", err)
		return
	}

	fmt.Printf("✅ 数据一致性校验完成:\n")
	if result, ok := resp.Data.(map[string]interface{}); ok {
		fmt.Printf("   校验时间: %v\n", result["validation_time"])
		fmt.Printf("   校验状态: %v\n", result["status"])
		fmt.Printf("   检查记录数: %v\n", result["checked_records"])
		fmt.Printf("   不一致数量: %v\n", result["inconsistencies"])
		fmt.Printf("   校验消息: %v\n", result["message"])
	}
	fmt.Println()
}

// 在实际微服务中的集成示例
type UserService struct {
	blockchainClient *blockchain.Client
}

// NewUserService 创建用户服务
func NewUserService(blockchainURL string) *UserService {
	return &UserService{
		blockchainClient: blockchain.NewClient(blockchainURL),
	}
}

// UpdateUserStatus 更新用户状态（集成区块链记录）
func (s *UserService) UpdateUserStatus(ctx context.Context, userID, versionSource, oldStatus, newStatus, reason, operatorID string) error {
	// 1. 更新数据库中的用户状态
	if err := s.updateDatabaseUserStatus(userID, newStatus); err != nil {
		return fmt.Errorf("更新数据库用户状态失败: %v", err)
	}

	// 2. 记录到区块链（异步，不影响主业务）
	go func() {
		blockchainReq := &blockchain.VersionStatusChangeRequest{
			UserID:        userID,
			VersionSource: versionSource,
			OldStatus:     oldStatus,
			NewStatus:     newStatus,
			ChangeReason:  reason,
			OperatorID:    operatorID,
			Remark:        "用户服务自动同步",
		}

		blockchainCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := s.blockchainClient.RecordVersionStatusChange(blockchainCtx, blockchainReq); err != nil {
			log.Printf("记录用户状态到区块链失败: %v", err)
			// 这里可以实现重试机制或保存到失败队列
		} else {
			log.Printf("用户状态已记录到区块链: userID=%s, %s->%s", userID, oldStatus, newStatus)
		}
	}()

	return nil
}

// updateDatabaseUserStatus 更新数据库中的用户状态（模拟）
func (s *UserService) updateDatabaseUserStatus(userID, newStatus string) error {
	// 这里实现实际的数据库更新逻辑
	log.Printf("更新数据库用户状态: userID=%s, status=%s", userID, newStatus)
	return nil
}
