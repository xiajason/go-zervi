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

func CreateResumeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateResumeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := resume.NewCreateResumeLogic(r.Context(), svcCtx)
		resp, err := l.CreateResume(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
