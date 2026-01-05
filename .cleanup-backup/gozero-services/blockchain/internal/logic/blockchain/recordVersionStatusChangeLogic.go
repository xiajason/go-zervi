// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blockchain

import (
	"context"

	"blockchain/internal/svc"
	"blockchain/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordVersionStatusChangeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecordVersionStatusChangeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordVersionStatusChangeLogic {
	return &RecordVersionStatusChangeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecordVersionStatusChangeLogic) RecordVersionStatusChange(req *types.VersionStatusChangeRequest) (resp *types.VersionStatusChangeResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
