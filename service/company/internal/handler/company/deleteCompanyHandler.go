// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"net/http"

	"company/internal/logic/company"
	"company/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteCompanyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := company.NewDeleteCompanyLogic(r.Context(), svcCtx)
		resp, err := l.DeleteCompany()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
