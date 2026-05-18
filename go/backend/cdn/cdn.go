package cdn

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
)

var key string

const CDNBaseURL = "https://cdn.fynbos.app"

func Init(cdnKey string) {
	key = cdnKey
}

type PutArgs struct {
	Data        []byte
	ContentType string
	Path        string
}

func Put(ctx context.Context, args PutArgs) error {
	if key == "" {
		return nil
	}

	u, err := url.JoinPath(CDNBaseURL, args.Path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(args.Data))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", args.ContentType)
	req.Header.Add("X-Fynbos-Auth-Key", key)
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
