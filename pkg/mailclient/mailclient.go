package mailclient

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/google/uuid"
)

// mail client types.
const (
	// ClientTypeSMTP represents mail clients using SMTP protocol to send emails.
	ClientTypeSMTP = "smtp"
	// ClientTypeSMTP represents mail clients using console to print emails.
	ClientTypeConsole = "console"
)

// Client is used to handle sending of emails.
//
//go:generate mockgen -package=mailclientmocks -source=$GOFILE -destination=./mocks/mailclient.go
type Client interface {
	Send(
		to []string,
		subject string,
		textTmpl *texttemplate.Template,
		htmlTmpl *htmltemplate.Template,
		tmplData any,
	) error
}

// BuildMail builds multi-line email body using MIME format.
func BuildMail(
	from string,
	to []string,
	subject string,
	textTmpl *texttemplate.Template,
	htmlTmpl *htmltemplate.Template,
	tmplData any,
	timeProvider timekeeper.Provider,
) ([]byte, error) {
	boundary := uuid.NewString()

	var mailBody bytes.Buffer
	mailBody.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\n", boundary))
	mailBody.WriteString("MIME-Version: 1.0\n")
	mailBody.WriteString(fmt.Sprintf("Subject: %s\n", subject))
	mailBody.WriteString(fmt.Sprintf("From: %s\n", from))
	mailBody.WriteString(fmt.Sprintf("From: %s\n", strings.Join(to, ", ")))
	mailBody.WriteString(fmt.Sprintf("Date: %s\n", timeProvider.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700")))

	mailBody.WriteString("\n")

	mailBody.WriteString(fmt.Sprintf("--%s\n", boundary))
	mailBody.WriteString("Content-Type: text/plain; charset=\"utf-8\"\n")
	mailBody.WriteString("MIME-Version: 1.0\n")

	mailBody.WriteString("\n")

	err := textTmpl.Execute(&mailBody, tmplData)
	if err != nil {
		return nil, errutils.FormatError(err, "textTmpl.Execute failed")
	}

	mailBody.WriteString("\n")

	mailBody.WriteString(fmt.Sprintf("--%s\n", boundary))
	mailBody.WriteString("Content-Type: text/html; charset=\"utf-8\"\n")
	mailBody.WriteString("MIME-Version: 1.0\n")

	mailBody.WriteString("\n")

	err = htmlTmpl.Execute(&mailBody, tmplData)
	if err != nil {
		return nil, errutils.FormatError(err, "htmlTmpl.Execute failed")
	}

	mailBody.WriteString("\n")

	mailBody.WriteString(fmt.Sprintf("--%s--\n", boundary))

	msg := mailBody.Bytes()

	return msg, nil
}
