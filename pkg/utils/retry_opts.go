package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type RetryOpts struct {
	broker   *nats.Conn
	Retries  int
	Timeout  time.Duration
	Subject  string
	Payload  []byte
}

func NewRetryOpts(b *nats.Conn) *RetryOpts {
	return &RetryOpts{broker: b, Retries: 10, Timeout: time.Second * 2}
}

func (r *RetryOpts) WithRetries(i int) *RetryOpts {
	r.Retries = i
	return r
}

func (r *RetryOpts) WithTimeout(t time.Duration) *RetryOpts {
	r.Timeout = t
	return r
}

func (r *RetryOpts) WithSubject(s string) *RetryOpts {
	r.Subject = s
	return r
}

func (r *RetryOpts) WithPayload(a []byte) *RetryOpts {
	r.Payload = a
	return r
}

func (r *RetryOpts) Run() (*nats.Msg, error) {
	if r.broker == nil {
		return nil, errors.New("No broker in RetryOpts")
	}

	if r.Retries == 0 {
		return nil, errors.New("No Retries in RetryOpts")
	}

	if r.Timeout == 0 {
		return nil, errors.New("No Timeout in RetryOpts")
	}

	if r.Subject == "" {
		return nil, errors.New("No Subject in RetryOpts")
	}

	for i := range r.Retries {
		resp, err := r.broker.Request(r.Subject, r.Payload, r.Timeout)
		if err != nil {
			if i+1 == r.Retries {
				return nil, err
			}

			fmt.Printf("%s %s %s", r.Subject, r.Payload, err)
			continue
		}

		return resp, nil
	}

	return nil, nil
}

