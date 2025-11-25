// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"context"

	"resume/internal/svc"
	"resume/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResumeLogic {
	return &GetResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetResumeLogic) GetResume() (resp *types.Resume, err error) {
	// todo: add your logic here and delete this line

	return
}
