// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"context"

	"ai/internal/svc"
	"ai/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SmartRecommendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSmartRecommendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SmartRecommendLogic {
	return &SmartRecommendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SmartRecommendLogic) SmartRecommend(req *types.SmartRecommendRequest) (resp *types.SmartRecommendResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
