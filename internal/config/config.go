package config

// Config represents config variables for the server.
type Config struct {
	Hostname             string `env:"NYMPHADORAAPI_HOSTNAME"`
	Port                 int    `env:"NYMPHADORAAPI_PORT"`
	SecretKey            string `env:"NYMPHADORAAPI_SECRET_KEY"`
	FrontendBaseURL      string `env:"NYMPHADORAAPI_FRONTEND_BASE_URL"`
	PostgresHostname     string `env:"NYMPHADORAAPI_POSTGRES_HOSTNAME"`
	PostgresPort         int    `env:"NYMPHADORAAPI_POSTGRES_PORT"`
	PostgresUsername     string `env:"NYMPHADORAAPI_POSTGRES_USERNAME"`
	PostgresPassword     string `env:"NYMPHADORAAPI_POSTGRES_PASSWORD"`
	PostgresDatabaseName string `env:"NYMPHADORAAPI_POSTGRES_DATABASE_NAME"`
	SMTPHostname         string `env:"NYMPHADORAAPI_SMTP_HOSTNAME"`
	SMTPPort             int    `env:"NYMPHADORAAPI_SMTP_PORT"`
	SMTPUsername         string `env:"NYMPHADORAAPI_SMTP_USERNAME"`
	SMTPPassword         string `env:"NYMPHADORAAPI_SMTP_PASSWORD"`
	MailClientType       string `env:"NYMPHADORAAPI_MAIL_CLIENT_TYPE"`
	PistonAPIKey         string `env:"NYMPHADORAAPI_PISTON_API_KEY"`
}
