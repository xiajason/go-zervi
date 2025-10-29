// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"

	"company/internal/svc"
	"company/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchCompaniesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchCompaniesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchCompaniesLogic {
	return &SearchCompaniesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchCompaniesLogic) SearchCompanies(req *types.CompanySearchRequest) (resp *types.CompanySearchResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
