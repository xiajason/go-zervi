// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"net/http"

	"ai/internal/logic/ai"
	"ai/internal/svc"
	"ai/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func MatchResumeJobHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AIMatchRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := ai.NewMatchResumeJobLogic(r.Context(), svcCtx)
		resp, err := l.MatchResumeJob(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
