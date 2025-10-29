// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"context"

	"resume/internal/svc"
	"resume/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateResumeLogic {
	return &UpdateResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateResumeLogic) UpdateResume(req *types.UpdateResumeRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
