// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"context"

	"ai/internal/svc"
	"ai/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAnalysisHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAnalysisHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAnalysisHistoryLogic {
	return &GetAnalysisHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAnalysisHistoryLogic) GetAnalysisHistory() (resp []types.ResumeAnalysisResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
