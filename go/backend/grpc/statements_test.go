package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/providers/unit"
	_user "gitlab.com/fynbos/backend/user"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetStatements(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	t.Run("Successfully calls GetStatements", func(t *testing.T) {
		userID := uuid.NewString()
		unitCustomerID := "41"
		unitAccountID := "451"
		c.UnitProvider.EXPECT().GetCustomerByIdentityID(gomock.Any(), userID).Return(&unit.Customer{
			ID: unitCustomerID,
		}, nil).Times(1)
		c.UnitProvider.EXPECT().GetStatements(gomock.Any(), unitCustomerID).Return([]unit.Statement{
			{
				ID:        "1",
				Period:    "2022-07",
				AccountID: unitAccountID,
			},
			{
				ID:        "2",
				Period:    "2022-08",
				AccountID: unitAccountID,
			},
		}, nil).Times(1)

		res, err := client.GetStatements(_user.ActingAsContext(t, context.Background(), &_user.User{
			ID: userID,
		}), &backendv1.Empty{})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(res.GetStatements()))
		assert.Equal(t, "1", res.GetStatements()[0].GetId())
		assert.Equal(t, "2", res.GetStatements()[1].GetId())
		assert.Equal(t, "2022-07", res.GetStatements()[0].GetPeriod())
		assert.Equal(t, unitAccountID, res.GetStatements()[0].GetAccountId())
	})

	t.Run("Successfully fails if no statements were found", func(t *testing.T) {
		c.UnitProvider.EXPECT().GetCustomerByIdentityID(gomock.Any(), gomock.Any()).Return(&unit.Customer{
			ID: uuid.NewString(),
		}, nil).Times(1)
		c.UnitProvider.EXPECT().GetStatements(gomock.Any(), gomock.Any()).Return(nil, unit.ErrNotFound).Times(1)

		res, err := client.GetStatements(_user.ActingAsContext(t, context.Background(), &_user.User{
			ID: uuid.NewString(),
		}), &backendv1.Empty{})

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.Equal(t, err, NotFoundError("No statements found."))
	})
}

func TestGetStatementPDF(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)

	t.Run("Successfully calls GetStatementPDF", func(t *testing.T) {
		statementID := "24"
		statementContentPDF := []byte("Statement PDF")
		c.UnitProvider.EXPECT().GetCustomerByIdentityID(gomock.Any(), gomock.Any()).Return(&unit.Customer{
			ID: "41",
		}, nil).Times(1)
		c.UnitProvider.EXPECT().GetStatementPDF(gomock.Any(), gomock.Any()).Return(&unit.StatementPDF{
			ID:  statementID,
			PDF: statementContentPDF,
		}, nil).Times(1)

		res, err := client.GetStatementPDF(_user.ActingAsContext(t, context.Background(), &_user.User{
			ID: uuid.NewString(),
		}), &backendv1.GetStatementPDFRequest{
			StatementId: statementID,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, statementID, res.GetId())
		assert.Equal(t, statementContentPDF, res.GetStatementPdf())
	})

	t.Run("Successfully handles erros from GetStatementPDF", func(t *testing.T) {
		c.UnitProvider.EXPECT().GetCustomerByIdentityID(gomock.Any(), gomock.Any()).Return(&unit.Customer{
			ID: uuid.NewString(),
		}, nil).Times(1)
		c.UnitProvider.EXPECT().GetStatementPDF(gomock.Any(), gomock.Any()).Return(nil, unit.ErrNotFound).Times(1)
		res, err := client.GetStatementPDF(_user.ActingAsContext(t, context.Background(), &_user.User{
			ID: uuid.NewString(),
		}), &backendv1.GetStatementPDFRequest{
			StatementId: "23",
		})

		assert.Nil(t, res)
		assert.Error(t, err)
		assert.Equal(t, err, NotFoundError("Failed to find statement."))
	})
}
