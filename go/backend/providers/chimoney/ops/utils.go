package ops

import (
	"fmt"
	"math"
	"strings"
)

func ExtractChiWalletIDFromIssueID(issueID string) (string, error) {
	parts := strings.Split(issueID, "_")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid issueID format: %s", issueID)
	}
	return parts[0], nil
}

func GetIntFromFloat64WithScale(amount float64, decimalPlaces int) int64 {
	return int64(amount * math.Pow10(decimalPlaces))
}
