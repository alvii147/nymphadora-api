package code_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/internal/code"
	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/validate"
	"github.com/stretchr/testify/require"
)

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

	testcases := []struct {
		name       string
		level      code.CodeSpaceAccessLevel
		wantString string
	}{
		{
			name:       "Read-only access level",
			level:      code.CodeSpaceAccessLevelReadOnly,
			wantString: api.CodeSpaceAccessLevelReadOnly,
		},
		{
			name:       "Read-write access level",
			level:      code.CodeSpaceAccessLevelReadWrite,
			wantString: api.CodeSpaceAccessLevelReadWrite,
		},
		{
			name:       "Unknown access level",
			level:      42,
			wantString: "",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testcase.wantString, testcase.level.String())
		})
	}
}

func TestGetAccessLevelFromString(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name        string
		accessLevel string
		wantLevel   code.CodeSpaceAccessLevel
	}{
		{
			name:        "Read-only access level",
			accessLevel: api.CodeSpaceAccessLevelReadOnly,
			wantLevel:   code.CodeSpaceAccessLevelReadOnly,
		},
		{
			name:        "Read-write access level",
			accessLevel: api.CodeSpaceAccessLevelReadWrite,
			wantLevel:   code.CodeSpaceAccessLevelReadWrite,
		},
		{
			name:        "Unknown access level",
			accessLevel: "DEADBEEF",
			wantLevel:   0,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testcase.wantLevel, code.GetAccessLevelFromString(testcase.accessLevel))
		})
	}
}
