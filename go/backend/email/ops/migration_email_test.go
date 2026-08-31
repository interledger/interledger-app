package ops

import (
	"context"
	"errors"
	"testing"

	"github.com/interledger/interledger-app/go/backend/email/sendgrid"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stretchr/testify/require"
)

type migrationSendgridClient struct {
	subject string
	to      []sendgrid.Email
	data    map[string]interface{}
	err     error
}

func (c *migrationSendgridClient) SendTemplate(_ context.Context, subject string, to []sendgrid.Email, _ string, data map[string]interface{}, _ []mail.Attachment) error {
	c.subject = subject
	c.to = append([]sendgrid.Email(nil), to...)
	c.data = data
	return c.err
}

func TestSendMigrationEmail(t *testing.T) {
	sg := &migrationSendgridClient{}
	b := &testBackends{
		external:       sg,
		applicationURL: "https://wallet.example",
	}

	paragraphs := []map[string]interface{}{
		{"heading": "What's changing"},
		{"paragraph": "We are migrating accounts."},
	}

	err := SendMigrationEmail(context.Background(), b, "Migration notice", "alice@example.com", "Alice", paragraphs)
	require.NoError(t, err)

	require.Equal(t, "Migration notice", sg.subject)
	require.Equal(t, []sendgrid.Email{{Name: "Alice", Address: "alice@example.com"}}, sg.to)
	require.Equal(t, "Migration notice", sg.data["subject"])

	data, ok := sg.data["data"].([]map[string]interface{})
	require.True(t, ok)
	require.Equal(t, map[string]interface{}{"paragraph": "Hello Alice,"}, data[0])
	require.Equal(t, paragraphs[0], data[1])
	require.Equal(t, paragraphs[1], data[2])

	cta, ok := sg.data["cta"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "View account", cta["text"])
	require.Equal(t, "https://wallet.example/login", cta["url"])
}

func TestSendMigrationEmailWithoutFirstName(t *testing.T) {
	sg := &migrationSendgridClient{}
	b := &testBackends{external: sg, applicationURL: "https://wallet.example/"}

	err := SendMigrationEmail(context.Background(), b, "Migration notice", "alice@example.com", "  ",
		[]map[string]interface{}{{"paragraph": "We are migrating accounts."}})
	require.NoError(t, err)

	data, ok := sg.data["data"].([]map[string]interface{})
	require.True(t, ok)
	require.Equal(t, map[string]interface{}{"paragraph": "Hello,"}, data[0])
	require.Equal(t, []sendgrid.Email{{Name: "", Address: "alice@example.com"}}, sg.to)

	cta, ok := sg.data["cta"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "https://wallet.example/login", cta["url"], "trailing slash must not double up")
}

func TestSendMigrationEmailErrors(t *testing.T) {
	paragraphs := []map[string]interface{}{{"paragraph": "We are migrating accounts."}}

	t.Run("sendgrid failure is returned", func(t *testing.T) {
		sendErr := errors.New("sendgrid down")
		b := &testBackends{external: &migrationSendgridClient{err: sendErr}, applicationURL: "https://wallet.example"}

		err := SendMigrationEmail(context.Background(), b, "Migration notice", "alice@example.com", "Alice", paragraphs)
		require.ErrorIs(t, err, sendErr)
	})

	t.Run("unusable application URL", func(t *testing.T) {
		sg := &migrationSendgridClient{}
		b := &testBackends{external: sg, applicationURL: "://nope"}

		err := SendMigrationEmail(context.Background(), b, "Migration notice", "alice@example.com", "Alice", paragraphs)
		require.Error(t, err)
		require.Nil(t, sg.data, "no email is sent when the CTA URL cannot be built")
	})
}
