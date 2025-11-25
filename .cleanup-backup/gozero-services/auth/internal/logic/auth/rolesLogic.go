// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"

	"auth/internal/svc"
	"auth/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RolesLogic {
	return &RolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RolesLogic) Roles() (resp []types.Role, err error) {
	// todo: add your logic here and delete this line

	return
}
