package external

type (
	User struct {
		Guid       string `json:"guid"`
		ID         string `json:"id"`
		IsDisabled bool   `json:"is_disabled"`
	}

	GetWidgetURLArgs struct {
		UserGuid                 string `json:"-"`
		CurrentMemberGuid        string `json:"currentMemberGuid,omitempty"`
		DisableInstitutionSearch bool   `json:"disableInstitutionSearch,omitempty"`
		IncludeTransactions      bool   `json:"includeTransactions,omitempty"`
		IncludeIdentity          bool   `json:"includeIdentity,omitempty"`
		Mode                     string `json:"mode"`
		WidgetType               string `json:"widget_type"`
	}

	WidgetURL struct {
		WidgetType string `json:"type"`
		URL        string `json:"url"`
		UserID     string `json:"user_id"`
	}

	GetWidgerURLResponse struct {
		WidgetURL WidgetURL `json:"widget_url"`
	}

	Pagination struct {
		CurrentPage  int `json:"current_page"`
		PerPage      int `json:"per_page"`
		TotalEntries int `json:"total_entries"`
		TotalPages   int `json:"total_pages"`
	}

	ListUsersResponse struct {
		Users      []User     `json:"users"`
		Pagination Pagination `json:"pagination"`
	}
)
