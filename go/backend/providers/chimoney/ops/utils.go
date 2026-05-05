package ops

import (
	"fmt"
	"strings"
)

func ExtractChiWalletIDFromIssueID(issueID string) (string, error) {
	parts := strings.Split(issueID, "_")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid issueID format: %s", issueID)
	}
	return parts[0], nil
}
