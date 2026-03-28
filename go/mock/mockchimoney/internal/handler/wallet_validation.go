package handler

import (
	"context"
	"strings"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

func requireTrimmedField(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false
	}

	return trimmed, true
}

func (h *Handler) getSubAccountByID(ctx context.Context, id string) (models.SubAccount, error) {
	return h.store.GetSubAccount(ctx, strings.TrimSpace(id))
}

func (h *Handler) ensureSubAccountExists(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return nil
	}

	_, err := h.getSubAccountByID(ctx, id)
	return err
}
