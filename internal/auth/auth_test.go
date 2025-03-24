package auth_test

import (
	"context"
	"testing"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetUserUUIDFromContext(t *testing.T) {
	t.Parallel()

	userUUID := uuid.NewString()

	testcases := map[string]struct {
		ctx          context.Context
		wantUserUUID string
		wantErr      bool
	}{
		"User UUID in context": {
			ctx:          context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, userUUID),
			wantUserUUID: userUUID,
			wantErr:      false,
		},
		"No user UUID in context": {
			ctx:          context.Background(),
			wantUserUUID: "",
			wantErr:      true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			userUUID, err := auth.GetUserUUIDFromContext(testcase.ctx)

			if testcase.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wantUserUUID, userUUID)
		})
	}
}
