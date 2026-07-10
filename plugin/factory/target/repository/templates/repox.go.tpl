{{.Header}}

// Package repox is the shared runtime of the generated repositories: the error
// vocabulary adapters map their backend errors onto, the connection bundle the
// per-schema factories pick an adapter from, and the list-request shape every
// repository List takes (raw AIP-160 filter, parsed by the filterx engines).
package repox

import (
	"errors"

	"gorm.io/gorm"
{{- if .GraphQLModule}}

	{{.GraphQLPkg}} "{{.GraphQLModule}}"
{{- end}}
)

// Error sentinels the repository interfaces speak. Callers translate them to
// their transport's codes (e.g. gRPC NotFound / AlreadyExists / Aborted /
// InvalidArgument) with errors.Is.
var (
	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
	ErrConflict        = errors.New("conflict")
	ErrInvalidArgument = errors.New("invalid argument")
)

// Conn bundles the live backend handles a repository factory can adapt.
type Conn struct {
	Gorm *gorm.DB
{{- if .GraphQLModule}}
	GraphQL *{{.GraphQLPkg}}.Service
{{- end}}
}

// ListInput is the request shape every repository List takes: AIP page
// controls, an order_by string, and a raw AIP-160 filter expression.
type ListInput struct {
	PageSize  int32
	PageToken string
	OrderBy   string
	Filter    string
}
