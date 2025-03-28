package api_test

import (
	"testing"
	"time"

	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/jsonutils"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/stretchr/testify/require"
)

func TestCreateUserRequestValidate(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		req               *api.CreateUserRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request": {
			req: &api.CreateUserRequest{
				Email:     testkit.GenerateFakeEmail(),
				Password:  testkit.GenerateFakePassword(),
				FirstName: testkit.MustGenerateRandomString(8, true, true, false),
				LastName:  testkit.MustGenerateRandomString(8, true, true, false),
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Invalid email": {
			req: &api.CreateUserRequest{
				Email:     "1nv4l1d3m41l",
				Password:  testkit.GenerateFakePassword(),
				FirstName: testkit.MustGenerateRandomString(8, true, true, false),
				LastName:  testkit.MustGenerateRandomString(8, true, true, false),
			},
			wantValid:         false,
			wantInvalidFields: []string{"email"},
		},
		"Blank email": {
			req: &api.CreateUserRequest{
				Email:     "",
				Password:  testkit.GenerateFakePassword(),
				FirstName: testkit.MustGenerateRandomString(8, true, true, false),
				LastName:  testkit.MustGenerateRandomString(8, true, true, false),
			},
			wantValid:         false,
			wantInvalidFields: []string{"email"},
		},
		"Blank password": {
			req: &api.CreateUserRequest{
				Email:     testkit.GenerateFakeEmail(),
				Password:  "",
				FirstName: testkit.MustGenerateRandomString(8, true, true, false),
				LastName:  testkit.MustGenerateRandomString(8, true, true, false),
			},
			wantValid:         false,
			wantInvalidFields: []string{"password"},
		},
		"Blank first name": {
			req: &api.CreateUserRequest{
				Email:     testkit.GenerateFakeEmail(),
				Password:  testkit.GenerateFakePassword(),
				FirstName: "",
				LastName:  testkit.MustGenerateRandomString(8, true, true, false),
			},
			wantValid:         false,
			wantInvalidFields: []string{"first_name"},
		},
		"Blank last name": {
			req: &api.CreateUserRequest{
				Email:     testkit.GenerateFakeEmail(),
				Password:  testkit.GenerateFakePassword(),
				FirstName: testkit.MustGenerateRandomString(8, true, true, false),
				LastName:  "",
			},
			wantValid:         false,
			wantInvalidFields: []string{"last_name"},
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

func TestActivateUserRequestValidate(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		req               *api.ActivateUserRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request": {
			req: &api.ActivateUserRequest{
				Token: testkit.MustGenerateRandomString(8, true, true, false),
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Blank token": {
			req: &api.ActivateUserRequest{
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

func TestUpdateUserRequestValidate(t *testing.T) {
	t.Parallel()

	firstName := testkit.MustGenerateRandomString(8, true, true, false)
	lastName := testkit.MustGenerateRandomString(8, true, true, false)
	blankString := ""

	testcases := map[string]struct {
		req               *api.UpdateUserRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request, both first and last names": {
			req: &api.UpdateUserRequest{
				FirstName: &firstName,
				LastName:  &lastName,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, only first name": {
			req: &api.UpdateUserRequest{
				FirstName: &firstName,
				LastName:  nil,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, only last name": {
			req: &api.UpdateUserRequest{
				FirstName: nil,
				LastName:  &lastName,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, neither first nor last name": {
			req: &api.UpdateUserRequest{
				FirstName: nil,
				LastName:  nil,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Blank first name": {
			req: &api.UpdateUserRequest{
				FirstName: &blankString,
				LastName:  nil,
			},
			wantValid:         false,
			wantInvalidFields: []string{"first_name"},
		},
		"Blank last name": {
			req: &api.UpdateUserRequest{
				FirstName: nil,
				LastName:  &blankString,
			},
			wantValid:         false,
			wantInvalidFields: []string{"last_name"},
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

func TestCreateTokenRequestValidate(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		req               *api.CreateTokenRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request": {
			req: &api.CreateTokenRequest{
				Email:    testkit.GenerateFakeEmail(),
				Password: testkit.GenerateFakePassword(),
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Invalid email": {
			req: &api.CreateTokenRequest{
				Email:    "1nv4l1d3m41l",
				Password: testkit.GenerateFakePassword(),
			},
			wantValid:         false,
			wantInvalidFields: []string{"email"},
		},
		"Blank email": {
			req: &api.CreateTokenRequest{
				Email:    "",
				Password: testkit.GenerateFakePassword(),
			},
			wantValid:         false,
			wantInvalidFields: []string{"email"},
		},
		"Blank password": {
			req: &api.CreateTokenRequest{
				Email:    testkit.GenerateFakeEmail(),
				Password: "",
			},
			wantValid:         false,
			wantInvalidFields: []string{"password"},
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

func TestRefreshTokenRequestValidate(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		req               *api.RefreshTokenRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request": {
			req: &api.RefreshTokenRequest{
				Refresh: testkit.MustGenerateRandomString(8, true, true, false),
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Blank refresh token": {
			req: &api.RefreshTokenRequest{
				Refresh: "",
			},
			wantValid:         false,
			wantInvalidFields: []string{"refresh"},
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

func TestValidateTokenRequestValidate(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		req               *api.ValidateTokenRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request": {
			req: &api.ValidateTokenRequest{
				Token: testkit.MustGenerateRandomString(8, true, true, false),
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Blank token": {
			req: &api.ValidateTokenRequest{
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

func TestCreateAPIKeyRequestValidate(t *testing.T) {
	t.Parallel()

	timeProvider := timekeeper.NewFrozenProvider()
	now := timeProvider.Now()

	testcases := map[string]struct {
		req               *api.CreateAPIKeyRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request, no expiry": {
			req: &api.CreateAPIKeyRequest{
				Name:      testkit.MustGenerateRandomString(8, true, true, false),
				ExpiresAt: nil,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, with expiry": {
			req: &api.CreateAPIKeyRequest{
				Name:      testkit.MustGenerateRandomString(8, true, true, false),
				ExpiresAt: &now,
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Blank name": {
			req: &api.CreateAPIKeyRequest{
				Name: "",
			},
			wantValid:         false,
			wantInvalidFields: []string{"name"},
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

func TestUpdateAPIKeyRequestValidate(t *testing.T) {
	t.Parallel()

	apiKeyName := testkit.MustGenerateRandomString(8, true, true, false)
	blankString := ""
	timeProvider := timekeeper.NewFrozenProvider()
	now := timeProvider.Now()

	testcases := map[string]struct {
		req               *api.UpdateAPIKeyRequest
		wantValid         bool
		wantInvalidFields []string
	}{
		"Valid request, with name, no expiry": {
			req: &api.UpdateAPIKeyRequest{
				Name: &apiKeyName,
				ExpiresAt: jsonutils.Optional[time.Time]{
					Valid: false,
					Value: nil,
				},
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, with name, null expiry": {
			req: &api.UpdateAPIKeyRequest{
				Name: &apiKeyName,
				ExpiresAt: jsonutils.Optional[time.Time]{
					Valid: true,
					Value: nil,
				},
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, with name and expiry": {
			req: &api.UpdateAPIKeyRequest{
				Name: &apiKeyName,
				ExpiresAt: jsonutils.Optional[time.Time]{
					Valid: true,
					Value: &now,
				},
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Valid request, no name, with expiry": {
			req: &api.UpdateAPIKeyRequest{
				Name: nil,
				ExpiresAt: jsonutils.Optional[time.Time]{
					Valid: true,
					Value: &now,
				},
			},
			wantValid:         true,
			wantInvalidFields: nil,
		},
		"Blank name": {
			req: &api.UpdateAPIKeyRequest{
				Name: &blankString,
				ExpiresAt: jsonutils.Optional[time.Time]{
					Valid: true,
					Value: &now,
				},
			},
			wantValid:         false,
			wantInvalidFields: []string{"name"},
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
