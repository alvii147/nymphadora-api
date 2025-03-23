package piston_test

import (
	"testing"

	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/alvii147/nymphadora-api/pkg/piston"
	"github.com/stretchr/testify/require"
)

const GoCode = `
package main

import "fmt"

func main() {
	fmt.Print("Hello, world!")
}
`

const PythonCode = `
print('Hello, world!', end='')
`

func TestPistonClientExecute(t *testing.T) {
	t.Parallel()

	client := piston.NewClient(nil, httputils.NewHTTPClient(nil))

	testcases := []struct {
		name     string
		fileName string
		language string
		version  string
		code     string
	}{
		{
			name:     "Go",
			fileName: "main.go",
			language: api.PistonLanguageGo,
			version:  api.PistonVersionGo,
			code:     GoCode,
		},
		{
			name:     "Python",
			fileName: "main.py",
			language: api.PistonLanguagePython,
			version:  api.PistonVersionPython,
			code:     PythonCode,
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

			require.Equal(t, "", response.Compile.Stdout)
			require.Equal(t, "", response.Compile.Stderr)
			require.Equal(t, "", response.Compile.Output)
			require.Nil(t, response.Compile.Code)
			require.Nil(t, response.Compile.Signal)

			require.Equal(t, "Hello, world!", response.Run.Stdout)
			require.Equal(t, "", response.Run.Stderr)
			require.Equal(t, "Hello, world!", response.Run.Output)
			require.Equal(t, 0, *response.Run.Code)
			require.Nil(t, response.Run.Signal)
		})
	}
}
