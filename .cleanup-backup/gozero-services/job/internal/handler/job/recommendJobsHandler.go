// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package job

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"job/internal/logic/job"
	"job/internal/svc"
	"job/internal/types"
)

func RecommendJobsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.JobRecommendRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := job.NewRecommendJobsLogic(r.Context(), svcCtx)
		resp, err := l.RecommendJobs(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
