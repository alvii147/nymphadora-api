package code_test

import (
	"os"
	"testing"

	"github.com/alvii147/nymphadora-api/internal/code"
	"github.com/alvii147/nymphadora-api/internal/database"
	"github.com/alvii147/nymphadora-api/internal/testkitinternal"
	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/validate"
	"github.com/stretchr/testify/require"
)

var TestDBPool database.Pool

func TestMain(m *testing.M) {
	TestDBPool = testkitinternal.MustNewDatabasePool()

	code := m.Run()

	TestDBPool.Close()
	os.Exit(code)
}

func TestCodeSpaceNameDataInitialized(t *testing.T) {
	t.Parallel()

	require.NotEmpty(t, code.Adjectives)
	require.NotEmpty(t, code.FileExtensions)
	require.NotEmpty(t, code.HarryPotterCharacters)
	require.NotEmpty(t, code.Pokemon)
	require.NotEmpty(t, code.ProgrammingTerms)
}

func TestCodeSpaceNameDataValidSlugs(t *testing.T) {
	t.Parallel()

	for _, word := range code.Adjectives {
		category := "adjective"
		t.Run(category+": "+word, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringSlug(category, word)
			require.True(t, v.Passed())
		})
	}

	for _, word := range code.FileExtensions {
		category := "file extension"
		t.Run(category+": "+word, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringSlug(category, word)
			require.True(t, v.Passed())
		})
	}

	for _, word := range code.HarryPotterCharacters {
		category := "harry potter character"
		t.Run(category+": "+word, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringSlug(category, word)
			require.True(t, v.Passed())
		})
	}

	for _, word := range code.Pokemon {
		category := "pokemon"
		t.Run(category+": "+word, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringSlug(category, word)
			require.True(t, v.Passed())
		})
	}

	for _, word := range code.ProgrammingTerms {
		category := "programming term"
		t.Run(category+": "+word, func(t *testing.T) {
			t.Parallel()

			v := validate.NewValidator()
			v.ValidateStringSlug(category, word)
			require.True(t, v.Passed())
		})
	}
}

func TestCodeSpaceAccessLevelString(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		level      code.CodeSpaceAccessLevel
		wantString string
	}{
		"Read-only access level": {
			level:      code.CodeSpaceAccessLevelReadOnly,
			wantString: api.CodeSpaceAccessLevelReadOnly,
		},
		"Read-write access level": {
			level:      code.CodeSpaceAccessLevelReadWrite,
			wantString: api.CodeSpaceAccessLevelReadWrite,
		},
		"Unknown access level": {
			level:      42,
			wantString: "",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testcase.wantString, testcase.level.String())
		})
	}
}

func TestGetAccessLevelFromString(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		accessLevel string
		wantLevel   code.CodeSpaceAccessLevel
	}{
		"Read-only access level": {
			accessLevel: api.CodeSpaceAccessLevelReadOnly,
			wantLevel:   code.CodeSpaceAccessLevelReadOnly,
		},
		"Read-write access level": {
			accessLevel: api.CodeSpaceAccessLevelReadWrite,
			wantLevel:   code.CodeSpaceAccessLevelReadWrite,
		},
		"Unknown access level": {
			accessLevel: "DEADBEEF",
			wantLevel:   0,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testcase.wantLevel, code.GetAccessLevelFromString(testcase.accessLevel))
		})
	}
}
