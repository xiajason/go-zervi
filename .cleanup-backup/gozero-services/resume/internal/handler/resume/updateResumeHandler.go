// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"resume/internal/logic/resume"
	"resume/internal/svc"
	"resume/internal/types"
)

func UpdateResumeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateResumeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := resume.NewUpdateResumeLogic(r.Context(), svcCtx)
		resp, err := l.UpdateResume(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
