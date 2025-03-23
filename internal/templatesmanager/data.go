package templatesmanager

const (
	// ActivationEmailSubject is the subject line of user activation emails.
	ActivationEmailSubject = "Welcome to Nymphadora!"
	// ActivationEmailTemplateName is the name of the user activation email template.
	ActivationEmailTemplateName = "activation"
	// CodeSpaceInvitationEmailSubject is the subject line of code space invitation emails.
	CodeSpaceInvitationEmailSubject = "You've been invited to collaborate on a code space!"
	// CodeSpaceInvitationEmailTemplateName is the name of the user activation email template.
	CodeSpaceInvitationEmailTemplateName = "codespaceinvitation"
)

// ActivationEmailTemplateData represents data for the user activation email template.
type ActivationEmailTemplateData struct {
	RecipientEmail string
	ActivationURL  string
}

// CodeSpaceInvitationEmailTemplateData represents data for the code space invitation email template.
type CodeSpaceInvitationEmailTemplateData struct {
	InvitationURL string
}
