// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"net/http"

	"ai/internal/logic/ai"
	"ai/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetMatchHistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := ai.NewGetMatchHistoryLogic(r.Context(), svcCtx)
		resp, err := l.GetMatchHistory()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
