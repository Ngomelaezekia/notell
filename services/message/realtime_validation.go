package main

import (
	"errors"
	"strings"
)

const messageMaxBytes = 10000

func validMessageBody(body string) error {
	if strings.TrimSpace(body) == "" {
		return errors.New("message body is required")
	}
	if len(body) > messageMaxBytes {
		return errors.New("message body is too large")
	}
	return nil
}
