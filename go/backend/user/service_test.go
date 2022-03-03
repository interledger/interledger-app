package user

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi"
	kratos "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/assert"
	test_utils "gitlab.com/fynbos/backend/utils"
	"go.uber.org/zap"
)

func TestAuthenticationService(s *testing.T) {
	// Adding this flag so that this test can be
	// disabled in CI. Our docker image doesn't have
	// docker-compose and so will fail in CI.
	kratosContainer, err := test_utils.SetupKratos()
	if err != nil {
		s.Fatal(err)
	}
	defer kratosContainer.Container.Terminate(context.Background())

	configuration := kratos.NewConfiguration()
	configuration.Servers = kratos.ServerConfigurations{
		{
			URL:         kratosContainer.URI,
			Description: "Dev Kratos",
		},
	}
	apiClient := kratos.NewAPIClient(configuration)

	logger, err := zap.NewDevelopment()
	if err != nil {
		s.Fatal(err)
	}
	defer logger.Sync()

	user, err := NewService(apiClient)
	if err != nil {
		s.Fatal(err)
	}
	user = NewLoggingService(user, logger)

	kratosCookie, identity, err := _testRegisterUser(context.Background(), kratosContainer.URI)
	if err != nil {
		s.Fatal(err)
	}

	router := chi.NewRouter()
	router.Use(MakeMiddleware(user))
	router.Handle("/whoami", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := user.ForContext(r.Context())
		if err != nil {
			http.Error(w, "No user found", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}))
	server := httptest.NewServer(router)
	defer server.Close()

	s.Run("can get session from Kratos", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/whoami", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(kratosCookie)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		user := User{}
		err = json.Unmarshal(body, &user)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, user.ID, identity.Id)
	})

	s.Run("returns Unauthorized if there is no cookie", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/whoami")
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("Returns Forbidden if cookie is invalid", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/whoami", nil)
		if err != nil {
			t.Fatal(err)
		}
		cookie := http.Cookie{
			Name:  "ory_kratos_session",
			Value: "test-cookie",
		}
		req.AddCookie(&cookie)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func _testRegisterUser(ctx context.Context, kratosUrl string) (*http.Cookie, kratos.Identity, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, kratos.Identity{}, err
	}
	client := http.Client{
		Jar: jar,
	}
	configuration := kratos.NewConfiguration()
	configuration.Servers = kratos.ServerConfigurations{
		{
			URL:         kratosUrl,
			Description: "Dev Kratos",
		},
	}
	configuration.HTTPClient = &client
	apiClient := kratos.NewAPIClient(configuration)

	flow, _, err := apiClient.V0alpha2Api.InitializeSelfServiceRegistrationFlowForBrowsers(ctx).Execute()
	if err != nil {
		return nil, kratos.Identity{}, err
	}

	csrfToken, ok := flow.Ui.Nodes[0].Attributes.UiNodeInputAttributes.Value.(string)
	if !ok {
		return nil, kratos.Identity{}, err
	}

	reg, _, err := apiClient.V0alpha2Api.
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
		return nil, kratos.Identity{}, err
	}

	url, err := url.Parse(kratosUrl)
	if err != nil {
		return nil, kratos.Identity{}, err
	}
	cookies := jar.Cookies(url)
	var kratosCookie *http.Cookie = nil
	for _, c := range cookies {
		if c.Name == "ory_kratos_session" {
			kratosCookie = c
		}
	}

	return kratosCookie, reg.Identity, nil
}
