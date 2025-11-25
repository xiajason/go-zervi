// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"context"

	"ai/internal/svc"
	"ai/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnalyzeResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnalyzeResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AnalyzeResumeLogic {
	return &AnalyzeResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnalyzeResumeLogic) AnalyzeResume(req *types.ResumeAnalysisRequest) (resp *types.ResumeAnalysisResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
