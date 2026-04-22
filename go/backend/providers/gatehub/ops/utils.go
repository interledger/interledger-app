package ops

import (
	"errors"
	"strconv"
	"strings"
)

// for new scale is 2
func StringToScaledUInt(input string) (int64, error) {
	input = strings.ReplaceAll(input, ",", "")
	if len(input) == 0 {
		return 0, errors.New("empty string provided")
	}

	parts := strings.Split(input, ".")
	if len(parts) < 1 || len(parts) > 2 {
		return 0, errors.New("invalid format should be numeric or numeric.numeric")
	}

	beforeDotValue, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid string before decimal point should be numeric")
	}

	var afterDotValue int64 = 0
	if len(parts) == 2 {
		afterDotValue, err = strconv.ParseInt(parts[1], 10, 64)
		if len(parts[1]) == 1 {
			afterDotValue = afterDotValue * 10
		}
		if err != nil {
			return 0, errors.New("invalid string after decimal point should be numeric")
		}
	}
	fee := int64(beforeDotValue*100 + afterDotValue)

	return fee, nil
}

func formatMaskedPan(pan string) string {
	last4 := pan[len(pan)-4:]
	return "**** " + last4
}
