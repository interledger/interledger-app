package ops

import (
	"net/http"
)

type EasyEquitiesRequest struct {
	WealthUserID int64  `json:"user_id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func GetIdentityHandler(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

	}
}
