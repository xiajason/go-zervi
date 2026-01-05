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

func RecordVersionStatusChangeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VersionStatusChangeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := blockchain.NewRecordVersionStatusChangeLogic(r.Context(), svcCtx)
		resp, err := l.RecordVersionStatusChange(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
