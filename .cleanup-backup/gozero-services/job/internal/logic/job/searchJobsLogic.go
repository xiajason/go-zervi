// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package job

import (
	"context"

	"job/internal/svc"
	"job/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchJobsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchJobsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchJobsLogic {
	return &SearchJobsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchJobsLogic) SearchJobs(req *types.JobSearchRequest) (resp *types.JobSearchResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
