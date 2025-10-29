// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blockchain

import (
	"net/http"

	"blockchain/internal/logic/blockchain"
	"blockchain/internal/svc"
	"blockchain/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ValidateDataConsistencyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DataConsistencyValidateRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := blockchain.NewValidateDataConsistencyLogic(r.Context(), svcCtx)
		resp, err := l.ValidateDataConsistency(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
