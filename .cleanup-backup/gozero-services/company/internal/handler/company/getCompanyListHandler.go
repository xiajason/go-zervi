// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"net/http"

	"company/internal/logic/company"
	"company/internal/svc"
	"company/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetCompanyListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CompanySearchRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := company.NewGetCompanyListLogic(r.Context(), svcCtx)
		resp, err := l.GetCompanyList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
