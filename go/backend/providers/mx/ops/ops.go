package ops

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/mx/external"
)

func GetWidget(ctx context.Context, b Backends, walletID string) (string, error) {
	externalUsers, err := b.External().ListUsersByID(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	var user *external.User
	for _, usr := range externalUsers.Users {
		if usr.ID == walletID {
			user = &usr
			break
		}
	}
	if user == nil {
		user, err = b.External().CreateUser(ctx, walletID)
		if err != nil {
			return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
		}
	}

	widget, err := b.External().GetWidgetURL(ctx, external.GetWidgetURLArgs{
		UserGuid:            user.Guid,
		IncludeTransactions: false,
		IncludeIdentity:     true,
		Mode:                "verification",
		WidgetType:          "connect_widget",
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return widget.URL, nil
}
