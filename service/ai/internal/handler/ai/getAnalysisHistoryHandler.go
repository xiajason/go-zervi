// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"net/http"

	"ai/internal/logic/ai"
	"ai/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAnalysisHistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := ai.NewGetAnalysisHistoryLogic(r.Context(), svcCtx)
		resp, err := l.GetAnalysisHistory()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
