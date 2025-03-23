package mailclient

import (
	"fmt"
	htmltemplate "html/template"
	"io"
	texttemplate "text/template"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
)

// consoleClient implements a Client that prints email contents to the console.
// This should typically be used in local development.
type consoleClient struct {
	username     string
	timeProvider timekeeper.Provider
	writer       io.Writer
}

// NewConsoleClient returns a new consoleClient.
func NewConsoleClient(username string, timeProvider timekeeper.Provider, writer io.Writer) *consoleClient {
	return &consoleClient{
		username:     username,
		timeProvider: timeProvider,
		writer:       writer,
	}
}

// Send prints email body to the console.
func (client *consoleClient) Send(
	to []string,
	subject string,
	textTmpl *texttemplate.Template,
	htmlTmpl *htmltemplate.Template,
	tmplData any,
) error {
	msg, err := BuildMail(
		client.username,
		to,
		subject,
		textTmpl,
		htmlTmpl,
		tmplData,
		client.timeProvider,
	)
	if err != nil {
		return errutils.FormatError(err)
	}

	fmt.Fprint(client.writer, string(msg))

	return nil
}
