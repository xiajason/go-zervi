// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"

	"company/internal/svc"
	"company/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCompanyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCompanyLogic {
	return &DeleteCompanyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCompanyLogic) DeleteCompany() (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
