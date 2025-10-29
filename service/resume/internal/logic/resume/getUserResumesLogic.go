// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"context"

	"resume/internal/svc"
	"resume/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserResumesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserResumesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserResumesLogic {
	return &GetUserResumesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserResumesLogic) GetUserResumes() (resp []types.Resume, err error) {
	// todo: add your logic here and delete this line

	return
}
