package rafiki

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func NewMockRafikiGraphqlServer(t *testing.T) *httptest.Server {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := r.Body.Close(); err != nil {
				t.Fatal(err)
			}
		}()

		query := string(body)
		if strings.Contains(query, "createAccount") {
			jsonStr := fmt.Sprintf(`{
				"data": {
					"createAccount": {
						"code": "200",
						"success": true,
						"message": "Created account",
						"account": {
							"id": "%s",
							"asset": {
								"code": "740",
								"scale": 2
							}
						}
					}
				}
			}`, uuid.NewString())

			_, err = w.Write([]byte(jsonStr))
			if err != nil {
				t.Fatal(err)
			}
			return
		}

		http.Error(w, "Unknown query/mutation.", 501)
	}))
	t.Cleanup(func() {
		svr.Close()
	})

	return svr
}
