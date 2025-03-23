package testkitinternal_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/internal/code"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMustCreateCodeSpaceSuccess(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	codeSpace, codeSpaceAccess := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	require.NotNil(t, codeSpace.AuthorUUID)
	require.Equal(t, author.UUID, *codeSpace.AuthorUUID)
	require.Equal(t, "python", codeSpace.Language)

	require.Equal(t, author.UUID, codeSpaceAccess.UserUUID)
	require.Equal(t, codeSpace.ID, codeSpaceAccess.CodeSpaceID)
	require.Equal(t, code.CodeSpaceAccessLevelReadWrite, codeSpaceAccess.Level)
}

func TestMustCreateCodeSpaceError(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		testkitinternal.MustCreateCodeSpace(t, uuid.NewString(), "python")
	})
}

func TestMustCreateCodeSpaceAccessSuccess(t *testing.T) {
	t.Parallel()

	author, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	codeSpace, _ := testkitinternal.MustCreateCodeSpace(t, author.UUID, "python")

	invitee, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	codeSpaceAccess := testkitinternal.MustCreateCodeSpaceAccess(
		t,
		invitee.UUID,
		codeSpace.ID,
		code.CodeSpaceAccessLevelReadOnly,
	)

	require.Equal(t, invitee.UUID, codeSpaceAccess.UserUUID)
	require.Equal(t, codeSpace.ID, codeSpaceAccess.CodeSpaceID)
	require.Equal(t, code.CodeSpaceAccessLevelReadOnly, codeSpaceAccess.Level)
}

func TestMustCreateCodeSpaceAccessError(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		testkitinternal.MustCreateCodeSpaceAccess(
			t,
			uuid.NewString(),
			314159265,
			code.CodeSpaceAccessLevelReadOnly,
		)
	})
}
