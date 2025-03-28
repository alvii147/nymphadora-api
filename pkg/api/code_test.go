package api_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateCodeSpaceRequestValidate(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		req               *api.CreateCodeSpaceRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request, language C": {
			req: &api.CreateCodeSpaceRequest{
				Language: api.PistonLanguageC,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, language C++": {
			req: &api.CreateCodeSpaceRequest{
				Language: api.PistonLanguageCPlusPlus,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, language Go": {
			req: &api.CreateCodeSpaceRequest{
				Language: api.PistonLanguageGo,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, language Java": {
			req: &api.CreateCodeSpaceRequest{
				Language: api.PistonLanguageJava,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, language JavaScript": {
			req: &api.CreateCodeSpaceRequest{
				Language: api.PistonLanguageJavaScript,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, language Python": {
			req: &api.CreateCodeSpaceRequest{
				Language: api.PistonLanguagePython,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, language Rust": {
			req: &api.CreateCodeSpaceRequest{
				Language: api.PistonLanguageRust,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, language TypeScript": {
			req: &api.CreateCodeSpaceRequest{
				Language: api.PistonLanguageTypeScript,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Unsupported language": {
			req: &api.CreateCodeSpaceRequest{
				Language: "un5upp0rt3dl4ngu4g3",
			},
			wantValid:         false,
			wantInvalidFields: []string{"language"},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			passed, failures := testcase.req.Validate()
			require.Equal(t, testcase.wantValid, passed)
			require.Len(t, failures, len(testcase.wantInvalidFields))

			for _, field := range testcase.wantInvalidFields {
				fieldFailures, ok := failures[field]
				require.True(t, ok)
				require.NotEmpty(t, fieldFailures)
			}
		})
	}
}

func TestUpdateCodeSpaceRequestValidate(t *testing.T) {
	t.Parallel()

	contents := "print('Hello')"

	testcases := map[string]struct {
		req               *api.UpdateCodeSpaceRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request, no contents": {
			req: &api.UpdateCodeSpaceRequest{
				Contents: nil,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, with contents": {
			req: &api.UpdateCodeSpaceRequest{
				Contents: &contents,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			passed, failures := testcase.req.Validate()
			require.Equal(t, testcase.wantValid, passed)
			require.Len(t, failures, len(testcase.wantInvalidFields))

			for _, field := range testcase.wantInvalidFields {
				fieldFailures, ok := failures[field]
				require.True(t, ok)
				require.NotEmpty(t, fieldFailures)
			}
		})
	}
}

func TestInviteCodeSpaceUserRequestValidate(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		req               *api.InviteCodeSpaceUserRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request, read-only access level": {
			req: &api.InviteCodeSpaceUserRequest{
				Email:       testkit.GenerateFakeEmail(),
				AccessLevel: "R",
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, read-write access level": {
			req: &api.InviteCodeSpaceUserRequest{
				Email:       testkit.GenerateFakeEmail(),
				AccessLevel: "W",
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Invalid email": {
			req: &api.InviteCodeSpaceUserRequest{
				Email:       "1nv4l1d3m41l",
				AccessLevel: "R",
			},
			wantValid:         false,
			wantInvalidFields: []string{"email"},
		},
		"Blank email": {
			req: &api.InviteCodeSpaceUserRequest{
				Email:       "",
				AccessLevel: "R",
			},
			wantValid:         false,
			wantInvalidFields: []string{"email"},
		},
		"Invalid access level": {
			req: &api.InviteCodeSpaceUserRequest{
				Email:       testkit.GenerateFakeEmail(),
				AccessLevel: "X",
			},
			wantValid:         false,
			wantInvalidFields: []string{"access_level"},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			passed, failures := testcase.req.Validate()
			require.Equal(t, testcase.wantValid, passed)
			require.Len(t, failures, len(testcase.wantInvalidFields))

			for _, field := range testcase.wantInvalidFields {
				fieldFailures, ok := failures[field]
				require.True(t, ok)
				require.NotEmpty(t, fieldFailures)
			}
		})
	}
}

func TestRemoveCodeSpaceUserRequestValidate(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		req               *api.RemoveCodeSpaceUserRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request": {
			req: &api.RemoveCodeSpaceUserRequest{
				UserUUID: uuid.NewString(),
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Blank user UUID": {
			req: &api.RemoveCodeSpaceUserRequest{
				UserUUID: "",
			},
			wantValid:         false,
			wantInvalidFields: []string{"user_uuid"},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			passed, failures := testcase.req.Validate()
			require.Equal(t, testcase.wantValid, passed)
			require.Len(t, failures, len(testcase.wantInvalidFields))

			for _, field := range testcase.wantInvalidFields {
				fieldFailures, ok := failures[field]
				require.True(t, ok)
				require.NotEmpty(t, fieldFailures)
			}
		})
	}
}

func TestAcceptCodeSpaceUserInvitationRequestValidate(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		req               *api.AcceptCodeSpaceUserInvitationRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request": {
			req: &api.AcceptCodeSpaceUserInvitationRequest{
				Token: testkit.MustGenerateRandomString(8, true, true, false),
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Blank token": {
			req: &api.AcceptCodeSpaceUserInvitationRequest{
				Token: "",
			},
			wantValid:         false,
			wantInvalidFields: []string{"token"},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			passed, failures := testcase.req.Validate()
			require.Equal(t, testcase.wantValid, passed)
			require.Len(t, failures, len(testcase.wantInvalidFields))

			for _, field := range testcase.wantInvalidFields {
				fieldFailures, ok := failures[field]
				require.True(t, ok)
				require.NotEmpty(t, fieldFailures)
			}
		})
	}
}
