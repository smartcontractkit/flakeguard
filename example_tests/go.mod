module github.com/smartcontractkit/flakeguard/example_tests

go 1.24.4

replace github.com/smartcontractkit/flakeguard => ../

require (
	github.com/smartcontractkit/flakeguard v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
