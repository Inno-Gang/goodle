package moodle

import "context"

type Authenticator[C any] interface {
	Authenticate(ctx context.Context, credentials C) (*Client, error)
}
