// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"context"

	"resume/internal/svc"
	"resume/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateResumeLogic {
	return &CreateResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateResumeLogic) CreateResume(req *types.CreateResumeRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
