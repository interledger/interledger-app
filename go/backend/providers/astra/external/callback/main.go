package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"gitlab.com/fynbos/backend/providers/astra/external"
)

func main() {

	_ = os.Setenv("ASTRA_CLIENT_ID", "29b899344bfb462d98fa4dff08ca1fe8")
	_ = os.Setenv("ASTRA_CLIENT_SECRET", "4c1c61166fe04bb686ae4dc8267d2e4c")

	time.Sleep(time.Second * 1)

	cl := external.New(nil)
	http.HandleFunc("/codeme", func(w http.ResponseWriter, req *http.Request) {

		body, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		fmt.Println("url", req.URL.String())

		fmt.Println("body", string(body))

		params, _ := url.ParseQuery(req.URL.RawQuery)

		newish, err := cl.CodeExchange(req.Context(), params.Get("code"))

		fmt.Println(newish)

		w.WriteHeader(http.StatusOK)
	})

	fmt.Println(http.ListenAndServe(":8080", nil))
}
