module github.com/DataDog/datadog-api-client-go/v2

go 1.22

retract (
	// Version used to retract v2.0.0 and v2.0.1. DO NOT USE.
	v2.0.1
	// Premature major version v2 release. DO NOT USE.
	v2.0.0
)

require (
	github.com/DataDog/zstd v1.5.2
	github.com/goccy/go-json v0.10.2
	github.com/google/uuid v1.5.0
	golang.org/x/oauth2 v0.10.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.17.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
