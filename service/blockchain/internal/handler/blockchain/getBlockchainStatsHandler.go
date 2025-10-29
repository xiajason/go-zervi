// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blockchain

import (
	"net/http"

	"blockchain/internal/logic/blockchain"
	"blockchain/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetBlockchainStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := blockchain.NewGetBlockchainStatsLogic(r.Context(), svcCtx)
		resp, err := l.GetBlockchainStats()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
