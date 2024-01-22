package external_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/providers/astra/external"
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
   "ssn":"123654789",
   "card_number":"12366998787",
   "sub":{
      "email":"maynard@tool.com",
      "ssn":"123654789",
      "card_number":"12366998787"
   }
}`,
			output: `{"card_number":"*****","email":"*****","ssn":"*****","sub":{"card_number":"*****","email":"*****","ssn":"*****"}}`,
		}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := external.Redact(context.Background(), []byte(tc.input))
			require.NoError(t, err)
			assert.Equal(t, tc.output, string(out))
		})
	}
}
