// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"context"

	"ai/internal/svc"
	"ai/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthCheckLogic {
	return &HealthCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthCheckLogic) HealthCheck() (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
