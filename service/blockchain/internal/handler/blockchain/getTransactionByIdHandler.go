// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blockchain

import (
	"net/http"

	"blockchain/internal/logic/blockchain"
	"blockchain/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetTransactionByIdHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := blockchain.NewGetTransactionByIdLogic(r.Context(), svcCtx)
		resp, err := l.GetTransactionById()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
