package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStatusFindDetail(t *testing.T) {
	t.Parallel()

	t.Run("returns detail when present", func(t *testing.T) {
		t.Parallel()
		br := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{Field: "email", Description: "invalid"},
			},
		}
		st := statusWithDetails(status.New(codes.InvalidArgument, "bad input"), br)

		got := statusFindDetail[*errdetails.BadRequest](st)
		require.NotNil(t, got)
		assert.Len(t, got.FieldViolations, 1)
		assert.Equal(t, "email", got.FieldViolations[0].Field)
	})

	t.Run("returns zero value when detail not present", func(t *testing.T) {
		t.Parallel()
		st := status.New(codes.InvalidArgument, "bad input")

		got := statusFindDetail[*errdetails.BadRequest](st)
		assert.Nil(t, got)
	})

	t.Run("returns first matching detail when multiple present", func(t *testing.T) {
		t.Parallel()
		first := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{Field: "first"},
			},
		}
		second := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{Field: "second"},
			},
		}
		st := statusWithDetails(status.New(codes.InvalidArgument, "bad input"), first, second)

		got := statusFindDetail[*errdetails.BadRequest](st)
		require.NotNil(t, got)
		assert.Equal(t, "first", got.FieldViolations[0].Field)
	})

	t.Run("returns correct type when multiple detail types present", func(t *testing.T) {
		t.Parallel()
		br := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{Field: "name"},
			},
		}
		ri := &errdetails.RetryInfo{}
		st := statusWithDetails(status.New(codes.InvalidArgument, "bad input"), ri, br)

		got := statusFindDetail[*errdetails.BadRequest](st)
		require.NotNil(t, got)
		assert.Equal(t, "name", got.FieldViolations[0].Field)

		gotRI := statusFindDetail[*errdetails.RetryInfo](st)
		require.NotNil(t, gotRI)
	})
}

func TestStatusWithDetails(t *testing.T) {
	t.Parallel()

	t.Run("attaches single detail", func(t *testing.T) {
		t.Parallel()
		br := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{Field: "email"},
			},
		}
		st := status.New(codes.InvalidArgument, "validation error")
		result := statusWithDetails(st, br)

		assert.Equal(t, codes.InvalidArgument, result.Code())
		assert.Equal(t, "validation error", result.Message())
		assert.Len(t, result.Details(), 1)
	})

	t.Run("attaches multiple details", func(t *testing.T) {
		t.Parallel()
		br := &errdetails.BadRequest{}
		ri := &errdetails.RetryInfo{}
		st := status.New(codes.Unavailable, "try again")
		result := statusWithDetails(st, br, ri)

		assert.Len(t, result.Details(), 2)
	})

	t.Run("preserves original status code and message", func(t *testing.T) {
		t.Parallel()
		st := status.New(codes.NotFound, "not found")
		result := statusWithDetails(st, &errdetails.BadRequest{})

		assert.Equal(t, codes.NotFound, result.Code())
		assert.Equal(t, "not found", result.Message())
	})

	t.Run("panics when it couldn't attach details", func(t *testing.T) {
		t.Parallel()
		st := status.New(codes.OK, "")

		assert.Panics(t, func() {
			statusWithDetails(st, &errdetails.BadRequest{})
		})
	})

}
