package templatesmanager

import (
	"embed"
	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/alvii147/nymphadora-api/pkg/errutils"
)

//go:embed email/*.txt email/*.html
var templatesFS embed.FS

// Manager manages and loads template files.
//
//go:generate mockgen -package=templatesmanagermocks -source=$GOFILE -destination=./mocks/manager.go
type Manager interface {
	Load(name string) (*texttemplate.Template, *htmltemplate.Template, error)
}

// manager implements Manager.
type manager struct{}

// NewManager creates and returns a new manager.
func NewManager() *manager {
	return &manager{}
}

// Load loads the text and html files by a given name.
func (m *manager) Load(name string) (*texttemplate.Template, *htmltemplate.Template, error) {
	textTmplFile := "email/" + name + ".txt"
	htmlTmplFile := "email/" + name + ".html"

	textTmpl, err := texttemplate.ParseFS(templatesFS, textTmplFile)
	if err != nil {
		return nil, nil, errutils.FormatErrorf(err, "text/template.ParseFS failed for file %s", textTmplFile)
	}

	htmlTmpl, err := htmltemplate.ParseFS(templatesFS, htmlTmplFile)
	if err != nil {
		return nil, nil, errutils.FormatErrorf(err, "html/template.ParseFS failed for file %s", htmlTmplFile)
	}

	return textTmpl, htmlTmpl, nil
}
