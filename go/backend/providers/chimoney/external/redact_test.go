package external_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/chimoney/external"
	"gotest.tools/assert"
)

func TestRedact(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		output string
	}{
		{
			name: "values",
			input: `{
   	"email":"maynard@tool.com",
   	"name":"Maynard Keenan",
   	"phoneNumber":"+1234556",
	"firstName":"Maynard"
}`,
			output: `{"email":"*****","firstName":"Maynard","name":"Maynard Keenan","phoneNumber":"*****"}`,
		}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := external.Redact(context.Background(), []byte(tc.input))
			require.NoError(t, err)
			assert.Equal(t, tc.output, string(out))
		})
	}
}
