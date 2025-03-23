package testkit_test

import (
	htmltemplate "html/template"
	"testing"
	texttemplate "text/template"

	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/stretchr/testify/require"
)

func TestMustParseMailMessageSuccess(t *testing.T) {
	t.Parallel()

	msg := `Content-Type: multipart/alternative; boundary="deadbeef"
MIME-Version: 1.0
Subject: 1L9cBLMEBzSn
From: vfgtd7ujt535@ucgufkizok.bih
From: y32y4v6iyx5i@lyijjmvasg.tcn
Date: Thu, 25 Jan 2024 14:11:10 +0000

--deadbeef
Content-Type: text/plain; charset="utf-8"
MIME-Version: 1.0

Text Message
--deadbeef
Content-Type: text/html; charset="utf-8"
MIME-Version: 1.0

HTML Message
--deadbeef--
	`

	textMsg, htmlMsg := testkit.MustParseMailMessage(msg)
	require.Contains(t, textMsg, "Text Message")
	require.NotContains(t, textMsg, "HTML Message")
	require.Contains(t, htmlMsg, "HTML Message")
	require.NotContains(t, htmlMsg, "Text Message")
}

func TestMustParseMailMessageError(t *testing.T) {
	t.Parallel()

	invalidMsg := "1nv4l1d m3554g3"
	msgWithInvalidContentType := `Content-Type: invalid/type; boundary="deadbeef"
MIME-Version: 1.0
Subject: 1L9cBLMEBzSn
From: vfgtd7ujt535@ucgufkizok.bih
From: y32y4v6iyx5i@lyijjmvasg.tcn
Date: Thu, 25 Jan 2024 14:11:10 +0000

--deadbeef
Content-Type: text/plain; charset="utf-8"
MIME-Version: 1.0

Text Message
--deadbeef
Content-Type: text/html; charset="utf-8"
MIME-Version: 1.0

HTML Message
--deadbeef--
	`
	msgWithOneSection := `Content-Type: multipart/alternative; boundary="deadbeef"
MIME-Version: 1.0
Subject: 1L9cBLMEBzSn
From: vfgtd7ujt535@ucgufkizok.bih
From: y32y4v6iyx5i@lyijjmvasg.tcn
Date: Thu, 25 Jan 2024 14:11:10 +0000

--deadbeef
Content-Type: text/plain; charset="utf-8"
MIME-Version: 1.0

Text Message
--deadbeef
	`

	testcases := []struct {
		name string
		msg  string
	}{
		{
			name: "Invalid message",
			msg:  invalidMsg,
		},
		{
			name: "Message with invalid content type",
			msg:  msgWithInvalidContentType,
		},
		{
			name: "Message with one section",
			msg:  msgWithOneSection,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			require.Panics(t, func() {
				testkit.MustParseMailMessage(testcase.msg)
			})
		})
	}
}

func TestInMemMailClientSend(t *testing.T) {
	t.Parallel()

	username := testkit.GenerateFakeEmail()
	timeProvider := timekeeper.NewFrozenProvider()
	client := testkit.NewInMemMailClient(username, timeProvider)

	mailCount := len(client.Logs)

	to := testkit.GenerateFakeEmail()
	subject := testkit.MustGenerateRandomString(12, true, true, true)
	textTmpl, err := texttemplate.New("textTmpl").Parse("Test Template Content: {{ .Value }}")
	require.NoError(t, err)
	htmlTmpl, err := htmltemplate.New("htmlTmpl").Parse("<div>Test Template Content: {{ .Value }}</div>")
	require.NoError(t, err)
	tmplData := map[string]int{
		"Value": 42,
	}

	err = client.Send([]string{to}, subject, textTmpl, htmlTmpl, tmplData)
	require.NoError(t, err)

	require.Len(t, client.Logs, mailCount+1)

	lastMail := client.Logs[len(client.Logs)-1]
	require.Equal(t, username, lastMail.From)
	require.Equal(t, []string{to}, lastMail.To)
	require.Equal(t, subject, lastMail.Subject)
	require.WithinDuration(t, timeProvider.Now(), lastMail.SentAt, testkit.TimeToleranceExact)

	textMsg, htmlMsg := testkit.MustParseMailMessage(string(lastMail.Message))

	require.Regexp(t, `Content-Type:\s*text\/plain;\s*charset\s*=\s*"utf-8"`, textMsg)
	require.Regexp(t, `MIME-Version:\s*1.0`, textMsg)
	require.Contains(t, textMsg, "Test Template Content: 42")

	require.Regexp(t, `Content-Type:\s*text\/html;\s*charset\s*=\s*"utf-8"`, htmlMsg)
	require.Regexp(t, `MIME-Version:\s*1.0`, htmlMsg)
	require.Contains(t, htmlMsg, "Test Template Content: 42")
}
