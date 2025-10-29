// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"

	"company/internal/svc"
	"company/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCompanyListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCompanyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanyListLogic {
	return &GetCompanyListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanyListLogic) GetCompanyList(req *types.CompanySearchRequest) (resp *types.CompanySearchResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
