// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"

	"company/internal/svc"
	"company/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCompanyStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCompanyStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanyStatsLogic {
	return &GetCompanyStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanyStatsLogic) GetCompanyStats() (resp *types.CompanyStats, err error) {
	// todo: add your logic here and delete this line

	return
}
