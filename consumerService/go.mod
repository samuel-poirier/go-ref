module github.com/samuel-poirier/go-ref/consumer

go 1.25.1

require (
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/samuel-poirier/go-ref/events v0.0.0-00010101000000-000000000000
	github.com/samuel-poirier/go-ref/shared v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.10.0
	github.com/swaggo/http-swagger/v2 v2.0.2
	github.com/swaggo/swag v1.16.6
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/swaggo/files/v2 v2.0.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/go-playground/validator/v10 v10.27.0
	github.com/golang-migrate/migrate/v4 v4.19.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/rabbitmq/amqp091-go v1.10.0
)

replace github.com/samuel-poirier/go-ref/shared => ../pkg/shared

replace github.com/samuel-poirier/go-ref/events => ../pkg/events
