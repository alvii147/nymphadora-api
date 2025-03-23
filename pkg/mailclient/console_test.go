package mailclient_test

import (
	"bytes"
	htmltemplate "html/template"
	"testing"
	texttemplate "text/template"

	"github.com/alvii147/nymphadora-api/pkg/mailclient"
	"github.com/alvii147/nymphadora-api/pkg/testkit"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
	"github.com/stretchr/testify/require"
)

func TestConsoleMailClientSend(t *testing.T) {
	t.Parallel()

	username := testkit.GenerateFakeEmail()
	var buf bytes.Buffer

	timeProvider := timekeeper.NewFrozenProvider()
	client := mailclient.NewConsoleClient(username, timeProvider, &buf)

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

	textMsg, htmlMsg := testkit.MustParseMailMessage(buf.String())

	require.Regexp(t, `Content-Type:\s*text\/plain;\s*charset\s*=\s*"utf-8"`, textMsg)
	require.Regexp(t, `MIME-Version:\s*1.0`, textMsg)
	require.Contains(t, textMsg, "Test Template Content: 42")

	require.Regexp(t, `Content-Type:\s*text\/html;\s*charset\s*=\s*"utf-8"`, htmlMsg)
	require.Regexp(t, `MIME-Version:\s*1.0`, htmlMsg)
	require.Contains(t, htmlMsg, "Test Template Content: 42")
}
