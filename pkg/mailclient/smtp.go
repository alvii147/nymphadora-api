package mailclient

import (
	"fmt"
	htmltemplate "html/template"
	"net/smtp"
	texttemplate "text/template"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
)

// smtpClient implements a Client that sends email through SMTP server.
// This should typically be used in production.
type smtpClient struct {
	hostname     string
	addr         string
	username     string
	password     string
	timeProvider timekeeper.Provider
}

// NewSMTPClient returns a new smtpClient.
func NewSMTPClient(
	hostname string,
	port int,
	username string,
	password string,
	timeProvider timekeeper.Provider,
) *smtpClient {
	return &smtpClient{
		hostname:     hostname,
		addr:         fmt.Sprintf("%s:%d", hostname, port),
		username:     username,
		password:     password,
		timeProvider: timeProvider,
	}
}

// Send sends an email through SMTP server.
func (client *smtpClient) Send(
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

	auth := smtp.PlainAuth("", client.username, client.password, client.hostname)
	err = smtp.SendMail(client.addr, auth, client.username, to, msg)
	if err != nil {
		return errutils.FormatError(err, "smtp.SendMail failed")
	}

	return nil
}
