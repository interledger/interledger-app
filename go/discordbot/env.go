package main

type Environment struct {
	DbURL           string `env:"DB_URL,default=cockroach://backend@cockroachdb-public:26257/backend?sslmode=verify-full&sslrootcert=/cockroach-certs/ca.crt&sslcert=/cockroach-certs/client.backend.crt&sslkey=/cockroach-certs/client.backend.key&max_conns=20&max_idle_conns=4"`
	KratosURL       string `env:"KRATOS_URL,default=http://localhost:4433"`
	KratosAdminURL  string `env:"KRATOS_ADMIN_URL,default=http://localhost:4433"`
	DiscordBotToken string `env:"DISCORD_BOT_TOKEN,required=true"`
}
