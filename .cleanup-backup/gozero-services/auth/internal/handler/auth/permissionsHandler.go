// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"net/http"

	"auth/internal/logic/auth"
	"auth/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := auth.NewPermissionsLogic(r.Context(), svcCtx)
		resp, err := l.Permissions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
