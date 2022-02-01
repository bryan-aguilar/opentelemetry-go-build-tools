module go.opentelemetry.io/build-tools/crosslink

go 1.17

require (
	github.com/google/go-cmp v0.5.7
	github.com/otiai10/copy v1.7.0
	github.com/spf13/cobra v1.3.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/build-tools v0.0.0-20220110194441-2a9d5288bd70
	go.uber.org/zap v1.20.0
	golang.org/x/mod v0.5.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace go.opentelemetry.io/build-tools => ../
