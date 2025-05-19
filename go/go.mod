module gitlab.com/fynbos

go 1.23.0

toolchain go1.23.6

require (
	github.com/AlecAivazis/survey/v2 v2.3.6
	github.com/Khan/genqlient v0.6.0
	github.com/aws/aws-sdk-go-v2/config v1.8.3
	github.com/aws/aws-sdk-go-v2/service/s3 v1.16.1
	github.com/bxcodec/faker/v3 v3.6.0
	github.com/cockroachdb/cockroach-go v2.0.1+incompatible
	github.com/coreos/go-oidc/v3 v3.1.0
	github.com/dgryski/trifles v0.0.0-20200323201526-dd97f9abfb48
	github.com/dunglas/httpsfv v1.0.1
	github.com/fogleman/gg v1.3.0
	github.com/getsentry/sentry-go v0.22.0
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-playground/validator/v10 v10.11.1
	github.com/golang-migrate/migrate/v4 v4.15.1
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/vault/api v1.9.0
	github.com/jackc/pgx/v5 v5.5.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/joho/godotenv v1.4.0
	github.com/lestrrat-go/jwx/v2 v2.0.18
	github.com/lib/pq v1.10.3
	github.com/machinebox/graphql v0.2.2
	github.com/n0madic/twitter-scraper v0.0.0-20230423143228-6d13e319a39d
	github.com/ory/kratos-client-go v0.13.1
	github.com/pusher/pusher-http-go/v5 v5.1.1
	github.com/riandyrn/otelchi v0.4.0
	github.com/segmentio/analytics-go/v3 v3.2.1
	github.com/sendgrid/rest v2.6.9+incompatible
	github.com/sendgrid/sendgrid-go v3.12.0+incompatible
	github.com/signintech/gopdf v0.15.1
	github.com/slack-go/slack v0.12.2
	github.com/smartystreets/smartystreets-go-sdk v1.15.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.10.0
	github.com/testcontainers/testcontainers-go v0.12.0
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.1.16
	github.com/uptrace/opentelemetry-go-extra/otelsqlx v0.1.16
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.36.4
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.36.4
	go.opentelemetry.io/otel v1.11.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.9.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.9.0
	go.opentelemetry.io/otel/sdk v1.9.0
	go.temporal.io/api v1.46.0
	go.temporal.io/sdk v1.34.0
	go.temporal.io/sdk/contrib/opentelemetry v0.2.0
	go.uber.org/zap v1.19.1
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240827150818-7e3bb234dfed
	google.golang.org/grpc v1.66.0
	google.golang.org/protobuf v1.36.5
	gopkg.in/stretchr/testify.v1 v1.2.2
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools v2.2.0+incompatible
)

require (
	github.com/aws/aws-sdk-go-v2 v1.9.2 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.6.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.2.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.7.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.4.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.7.2 // indirect
	github.com/aws/smithy-go v1.8.0 // indirect
	github.com/cenkalti/backoff/v3 v3.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cockroachdb/cockroach-go/v2 v2.2.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.6 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.4 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/phpdave11/gofpdi v1.0.13 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/segmentio/backo-go v1.0.0 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vektah/gqlparser/v2 v2.5.11 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.opentelemetry.io/contrib v1.0.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.9.0 // indirect
	go.opentelemetry.io/otel/metric v0.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.11.1 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	golang.org/x/image v0.0.0-20220302094943-723b81ca9867 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/term v0.29.0 // indirect
	google.golang.org/genproto v0.0.0-20230815205213-6bfd019c3878 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240827150818-7e3bb234dfed // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Basis-Theory/basistheory-go/v3 v3.13.0
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/Microsoft/hcsshim v0.8.21 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/containerd/cgroups v1.0.1 // indirect
	github.com/containerd/containerd v1.5.7 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.11+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/gogo/googleapis v1.4.1
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gogo/status v1.1.1
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/matryer/is v1.4.0 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.5.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v1.0.2 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/twilio/twilio-go v1.2.1
	github.com/urfave/cli/v2 v2.24.4
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/net v0.36.0
	golang.org/x/oauth2 v0.22.0
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
