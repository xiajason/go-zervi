// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blockchain

import (
	"context"

	"blockchain/internal/svc"
	"blockchain/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordTransactionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecordTransactionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordTransactionLogic {
	return &RecordTransactionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecordTransactionLogic) RecordTransaction(req *types.RecordTransactionRequest) (resp *types.RecordTransactionResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
