package testkit

import (
	htmltemplate "html/template"
	"io"
	"net/mail"
	"regexp"
	"strings"
	texttemplate "text/template"
	"time"
	"unicode/utf8"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/mailclient"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
)

// MustParseMailMessage parses a given mail body and panics on error.
func MustParseMailMessage(msg string) (string, string) {
	mailMsg, err := mail.ReadMessage(strings.NewReader(msg))
	if err != nil {
		panic(errutils.FormatError(err, "mail.ReadMessage failed"))
	}

	r := regexp.MustCompile(`^multipart/alternative;\s*boundary="(\S+)"$`)
	matches := r.FindStringSubmatch(mailMsg.Header.Get("Content-Type"))
	if len(matches) != 2 {
		panic("MustParseMailMessage failed to find content type")
	}

	boundary := matches[1]

	msgBytes, err := io.ReadAll(mailMsg.Body)
	if err != nil {
		panic(errutils.FormatError(err, "io.ReadAll failed"))
	}

	r = regexp.MustCompile(`--+` + boundary + `-*`)
	mailSections := r.Split(string(msgBytes), -1)

	nonEmptyMailSections := []string{}
	for _, sec := range mailSections {
		if utf8.RuneCountInString(strings.TrimSpace(sec)) != 0 {
			nonEmptyMailSections = append(nonEmptyMailSections, sec)
		}
	}

	if len(nonEmptyMailSections) != 2 {
		panic(errutils.FormatError(nil, "parsing text and html sections failed"))
	}

	textMsg := nonEmptyMailSections[0]
	htmlMsg := nonEmptyMailSections[1]

	return textMsg, htmlMsg
}

// inMemMailLogEntry represents an in-memory entry of an email event.
type inMemMailLogEntry struct {
	From    string
	To      []string
	Subject string
	Message []byte
	SentAt  time.Time
}

// inMemMailClient implements a Client that saves email data in local memory.
// This should typically be used in unit tests.
type inMemMailClient struct {
	username     string
	timeProvider timekeeper.Provider
	Logs         []inMemMailLogEntry
}

// NewInMemMailClient returns a new inMemMailClient.
func NewInMemMailClient(username string, timeProvider timekeeper.Provider) *inMemMailClient {
	return &inMemMailClient{
		username:     username,
		timeProvider: timeProvider,
		Logs:         make([]inMemMailLogEntry, 0),
	}
}

// Send adds an email event to in-memory storage.
func (client *inMemMailClient) Send(
	to []string,
	subject string,
	textTmpl *texttemplate.Template,
	htmlTmpl *htmltemplate.Template,
	tmplData any,
) error {
	msg, err := mailclient.BuildMail(
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

	client.Logs = append(
		client.Logs,
		inMemMailLogEntry{
			From:    client.username,
			To:      to,
			Subject: subject,
			Message: msg,
			SentAt:  client.timeProvider.Now(),
		},
	)

	return nil
}
