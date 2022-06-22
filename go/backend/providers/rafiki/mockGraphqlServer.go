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

		if strings.Contains(query, "createQuote") {
			fmt.Println(query)
			jsonStr := fmt.Sprintf(`{
				"data": {
					"createQuote": {
						"code": "200",
						"success": true,
						"message": "Created quote",
						"quote": {
							"id": "%s",
							"accountId": "",
							"receiver": "$ilp.test/receive",
							"sendAmount": {
								"value": 100,
								"assetCode": "740",
								"assetScale": 2
							},
							"receiveAmount": {
								"value": 99,
								"assetCode": "740",
								"assetScale": 2
							},
							"maxPacketAmount": 100,
							"minExchangeRate": 1.00,
							"lowEstimatedExchangeRate": 1.00,
							"highEstimatedExchangeRate": 1.00
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
