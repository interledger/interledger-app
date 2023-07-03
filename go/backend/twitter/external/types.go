package external

type (
	CreateTokenArgs struct {
		AuthCode     string
		CodeVerifier string
	}

	Tweet struct {
		Data struct {
			ID       string `json:"id"`
			Text     string `json:"text"`
			AuthorID string `json:"author_id"`
			Entities struct {
				URLs []struct {
					URL         string `json:"url"`
					ExpandedURL string `json:"expanded_url"`
				} `json:"urls"`
			} `json:"entities"`
		} `json:"data"`
		Includes struct {
			Users []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Username string `json:"username"`
			} `json:"users"`
		} `json:"includes"`
	}
)
