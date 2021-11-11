package user

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"time"

	kratos "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/backend/utils"
)

func TestAuthenticationService(s *testing.T) {
	identifier, err := test_utils.SetupKratos()
	if err != nil {
		s.Fatal(err)
	}
	defer test_utils.TeardownKratos(identifier)
	// wait for docker compose env to spin up.
	time.Sleep(2 * time.Second)

	jar, err := cookiejar.New(nil)
	if err != nil {
		s.Fatal(err)
	}
	client := http.Client{
		Jar: jar,
	}
	configuration := kratos.NewConfiguration()
	configuration.Servers = kratos.ServerConfigurations{
		{
			URL:         "http://127.0.0.1:4433",
			Description: "Dev Kratos",
		},
	}
	configuration.HTTPClient = &client
	apiClient := kratos.NewAPIClient(configuration)

	user, err := NewService(apiClient)
	if err != nil {
		s.Fatal(err)
	}

	s.Run("can get session from Kratos", func(t *testing.T) {
		ctx := context.Background()
		flow, _, err := apiClient.V0alpha2Api.InitializeSelfServiceRegistrationFlowForBrowsers(ctx).Execute()
		if err != nil {
			t.Fatal(err)
		}

		csrfToken, ok := flow.Ui.Nodes[0].Attributes.UiNodeInputAttributes.Value.(string)
		if !ok {
			t.Fatal("Could not get csrf token.")
		}

		_, _, err = apiClient.V0alpha2Api.
			SubmitSelfServiceRegistrationFlow(ctx).
			Flow(flow.GetId()).
			SubmitSelfServiceRegistrationFlowBody(kratos.SubmitSelfServiceRegistrationFlowBody{
				SubmitSelfServiceRegistrationFlowWithPasswordMethodBody: &kratos.SubmitSelfServiceRegistrationFlowWithPasswordMethodBody{
					Password: "testing*&1!",
					Method:   "password",
					Traits: map[string]interface{}{
						"email": "test@fynbos.dev",
					},
					CsrfToken: &csrfToken,
				},
			}).
			Execute()
		if err != nil {
			t.Fatal(err)
		}

		url, err := url.Parse("http://127.0.0.1:4433")
		if err != nil {
			t.Fatal(err)
		}
		cookies := jar.Cookies(url)
		// clear jar so we are sure the session cookie isn't being sent automatically.
		jar.SetCookies(url, []*http.Cookie{})

		user, err := user.GetUser(cookies[0])
		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, user)
	})

	s.Run("returns nil if no cookie is supplied", func(t *testing.T) {
		user, err := user.GetUser(nil)
		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, user)
	})
}
