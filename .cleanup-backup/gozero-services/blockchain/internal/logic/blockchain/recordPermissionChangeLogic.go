// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blockchain

import (
	"context"

	"blockchain/internal/svc"
	"blockchain/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordPermissionChangeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecordPermissionChangeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordPermissionChangeLogic {
	return &RecordPermissionChangeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecordPermissionChangeLogic) RecordPermissionChange(req *types.PermissionChangeRequest) (resp *types.PermissionChangeResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
