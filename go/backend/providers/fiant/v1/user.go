package v1

import (
	"context"
	"net/http"
)

type userHandler struct {
	path string
	ctrl *Controller
}

/*
 note(bradu): this method is a stub and needs to be expanded
 this was used for initial testing of the new fiant integration
 the method should be updated to maybe parse the response (debatable)
 and return a list of users or an appropriate data structure
*/

// https://developers.platform.fiant.io/reference/getlistofusers
func (uh *userHandler) ListAll(ctx context.Context) (*http.Response, error) {
	path := uh.path // "users"
	return uh.ctrl.get(ctx, path)
}
