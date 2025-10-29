// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blockchain

import (
	"context"

	"blockchain/internal/svc"
	"blockchain/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlockchainStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBlockchainStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockchainStatsLogic {
	return &GetBlockchainStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlockchainStatsLogic) GetBlockchainStats() (resp *types.BlockchainStats, err error) {
	// todo: add your logic here and delete this line

	return
}
