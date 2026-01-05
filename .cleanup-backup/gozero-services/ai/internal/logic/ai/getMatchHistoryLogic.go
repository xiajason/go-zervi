// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"context"

	"ai/internal/svc"
	"ai/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMatchHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMatchHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMatchHistoryLogic {
	return &GetMatchHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMatchHistoryLogic) GetMatchHistory() (resp []types.AIMatchResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
