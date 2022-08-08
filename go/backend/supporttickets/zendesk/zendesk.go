package zendesk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const baseURL = "https://fynbos.zendesk.com/api/v2"

func NewClient(username, apiToken string) Client {
	return &client{apiToken: apiToken, userName: username}
}

type client struct {
	userName   string
	apiToken   string
	httpClient http.Client
}

func (c client) CreateTicket(ctx context.Context, email, name, description string) error {
	body := CreateTicketReq{
		Ticket: Ticket{
			Requester: Requester{
				Name:  name,
				Email: email,
			},
			Comment: Comment{Body: description},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/tickets", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.userName+"/token", c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	// Created successfully
	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("failed to create new ticket. status: %s body: %s", resp.Status, string(respBody))
}
