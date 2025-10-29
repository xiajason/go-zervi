// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package job

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"job/internal/logic/job"
	"job/internal/svc"
)

func DeleteJobHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := job.NewDeleteJobLogic(r.Context(), svcCtx)
		resp, err := l.DeleteJob()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
