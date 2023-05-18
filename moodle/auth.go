package moodle

type Authenticator[C any] interface {
	Authenticate(credentials C) (*Client, error)
}
