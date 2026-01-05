// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"resume/internal/logic/resume"
	"resume/internal/svc"
)

func GetResumeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := resume.NewGetResumeLogic(r.Context(), svcCtx)
		resp, err := l.GetResume()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
