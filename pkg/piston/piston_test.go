package piston_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/alvii147/nymphadora-api/pkg/piston"
	"github.com/stretchr/testify/require"
)

const CCode = `
#include <stdio.h>

int main() {
    printf("Hello, world!");

    return 0;
}
`

const PythonCode = `
print('Hello, world!', end='')
`

func TestPistonClientExecute(t *testing.T) {
	t.Parallel()

	client := piston.NewClient(nil, httputils.NewHTTPClient(nil))

	testcases := []struct {
		name               string
		fileName           string
		language           string
		version            string
		code               string
		wantCompileResults bool
	}{
		{
			name:               "Execute C code, expect compilation results",
			fileName:           "main.c",
			language:           api.PistonLanguageC,
			version:            api.PistonVersionC,
			code:               CCode,
			wantCompileResults: true,
		},
		{
			name:               "Execute Python code, expect only runtime results",
			fileName:           "main.py",
			language:           api.PistonLanguagePython,
			version:            api.PistonVersionPython,
			code:               PythonCode,
			wantCompileResults: false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			fileEncoding := "utf8"
			req := &api.PistonExecuteRequest{
				Language: testcase.language,
				Version:  testcase.version,
				Files: []api.PistonFile{
					{
						Name:     &testcase.fileName,
						Content:  testcase.code,
						Encoding: &fileEncoding,
					},
				},
			}

			response, err := client.Execute(req)

			require.NoError(t, err)
			require.Equal(t, testcase.language, response.Language)
			require.Equal(t, testcase.version, response.Version)

			require.Equal(t, "Hello, world!", response.Run.Stdout)
			require.Empty(t, response.Run.Stderr)
			require.Equal(t, "Hello, world!", response.Run.Output)
			require.NotNil(t, response.Run.Code)
			require.Equal(t, 0, *response.Run.Code)
			require.Nil(t, response.Run.Signal)

			if !testcase.wantCompileResults {
				require.Nil(t, response.Compile)

				return
			}

			require.NotNil(t, response.Compile)
			require.Empty(t, response.Compile.Stdout)
			require.Empty(t, response.Compile.Stderr)
			require.Empty(t, response.Compile.Output)
			require.NotNil(t, response.Compile.Code)
			require.Equal(t, 0, *response.Compile.Code)
			require.Nil(t, response.Compile.Signal)
		})
	}
}
