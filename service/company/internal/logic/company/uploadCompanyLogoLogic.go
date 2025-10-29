// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"

	"company/internal/svc"
	"company/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadCompanyLogoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadCompanyLogoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadCompanyLogoLogic {
	return &UploadCompanyLogoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadCompanyLogoLogic) UploadCompanyLogo() (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
