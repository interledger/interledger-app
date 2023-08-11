package persona

import "errors"

var ErrNotFound = errors.New("persona webhook: not found.")
var ErrIdempotencyDuplicate = errors.New("persona: idempotency duplicate.")
