package jobs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestSendMigrationEmailJob(t *testing.T) {
	recipients := make([]MigrationEmailRecipient, 0, 25)
	for i := 1; i <= 25; i++ { // more than migrationEmailConcurrency, so batching runs
		recipients = append(recipients, MigrationEmailRecipient{
			Email:     fmt.Sprintf("u%d@example.com", i),
			FirstName: fmt.Sprintf("User%d", i),
		})
	}

	params := migrationParams(SendMigrationEmailParams{Region: "ALL"})

	newEnv := func(t *testing.T) (*testsuite.TestWorkflowEnvironment, *Activity) {
		t.Helper()
		var suite testsuite.WorkflowTestSuite
		env := suite.NewTestWorkflowEnvironment()
		a := &Activity{}
		env.RegisterActivity(a.ListMigrationEmailRecipients)
		env.RegisterActivity(a.SendMigrationEmailToRecipient)
		return env, a
	}

	t.Run("emails every recipient and reports none failed", func(t *testing.T) {
		env, a := newEnv(t)
		env.OnActivity(a.ListMigrationEmailRecipients, mock.Anything, params).Return(recipients, nil)
		env.OnActivity(a.SendMigrationEmailToRecipient, mock.Anything, "Migration", mock.Anything, mock.Anything, params.Paragraphs).
			Return(nil)

		env.ExecuteWorkflow(SendMigrationEmailJob, params)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		var failed []string
		require.NoError(t, env.GetWorkflowResult(&failed))
		require.Empty(t, failed)
		env.AssertNumberOfCalls(t, "SendMigrationEmailToRecipient", len(recipients))
	})

	t.Run("a failed send is reported, and never retried", func(t *testing.T) {
		env, a := newEnv(t)
		env.OnActivity(a.ListMigrationEmailRecipients, mock.Anything, params).Return(recipients, nil)
		env.OnActivity(a.SendMigrationEmailToRecipient, mock.Anything, mock.Anything, "u3@example.com", mock.Anything, mock.Anything).
			Return(errors.New("sendgrid timeout"))
		env.OnActivity(a.SendMigrationEmailToRecipient, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)

		env.ExecuteWorkflow(SendMigrationEmailJob, params)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
		var failed []string
		require.NoError(t, env.GetWorkflowResult(&failed))
		require.Equal(t, []string{"u3@example.com"}, failed)
		// One call per recipient: a retry here would send a second email.
		env.AssertNumberOfCalls(t, "SendMigrationEmailToRecipient", len(recipients))
	})

	t.Run("invalid params fail before any listing or send", func(t *testing.T) {
		env, a := newEnv(t)
		env.OnActivity(a.ListMigrationEmailRecipients, mock.Anything, mock.Anything).Return(recipients, nil)

		env.ExecuteWorkflow(SendMigrationEmailJob, SendMigrationEmailParams{Paragraphs: params.Paragraphs, Region: "US"})

		require.True(t, env.IsWorkflowCompleted())
		require.ErrorContains(t, env.GetWorkflowError(), "subject is required")
		env.AssertNumberOfCalls(t, "ListMigrationEmailRecipients", 0)
		env.AssertNumberOfCalls(t, "SendMigrationEmailToRecipient", 0)
	})

	t.Run("a failed listing fails the job", func(t *testing.T) {
		env, a := newEnv(t)
		env.OnActivity(a.ListMigrationEmailRecipients, mock.Anything, params).
			Return(nil, errors.New("no user found for: ghost@example.com"))

		env.ExecuteWorkflow(SendMigrationEmailJob, params)

		require.True(t, env.IsWorkflowCompleted())
		require.ErrorContains(t, env.GetWorkflowError(), "no user found for: ghost@example.com")
		env.AssertNumberOfCalls(t, "SendMigrationEmailToRecipient", 0)
	})
}
