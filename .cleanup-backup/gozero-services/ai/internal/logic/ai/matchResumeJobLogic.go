// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"context"

	"ai/internal/svc"
	"ai/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MatchResumeJobLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMatchResumeJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MatchResumeJobLogic {
	return &MatchResumeJobLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MatchResumeJobLogic) MatchResumeJob(req *types.AIMatchRequest) (resp *types.AIMatchResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
