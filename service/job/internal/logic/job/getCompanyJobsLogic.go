// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package job

import (
	"context"

	"job/internal/svc"
	"job/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCompanyJobsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCompanyJobsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanyJobsLogic {
	return &GetCompanyJobsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanyJobsLogic) GetCompanyJobs(req *types.JobListRequest) (resp *types.JobSearchResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
