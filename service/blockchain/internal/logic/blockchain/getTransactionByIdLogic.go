// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blockchain

import (
	"context"

	"blockchain/internal/svc"
	"blockchain/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTransactionByIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTransactionByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTransactionByIdLogic {
	return &GetTransactionByIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTransactionByIdLogic) GetTransactionById() (resp *types.BlockchainTransaction, err error) {
	// todo: add your logic here and delete this line

	return
}
