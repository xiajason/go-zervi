// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blockchain

import (
	"context"

	"blockchain/internal/svc"
	"blockchain/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryTransactionHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryTransactionHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryTransactionHistoryLogic {
	return &QueryTransactionHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryTransactionHistoryLogic) QueryTransactionHistory(req *types.QueryTransactionHistoryRequest) (resp *types.QueryTransactionHistoryResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
