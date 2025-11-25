// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"resume/internal/logic/resume"
	"resume/internal/svc"
)

func UploadResumeFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := resume.NewUploadResumeFileLogic(r.Context(), svcCtx)
		resp, err := l.UploadResumeFile()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
