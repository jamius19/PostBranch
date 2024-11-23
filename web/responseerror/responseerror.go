package responseerror

import (
	"context"
	"errors"
	"strings"
)

const ErrorsContextKey = "requestErrors"

type ResponseErrors struct {
	Errors []string
}

func NewRequestErrors() *ResponseErrors {
	return &ResponseErrors{}
}

func (re *ResponseErrors) Add(error string) {
	re.Errors = append(re.Errors, error)
}

func AddError(ctx context.Context, error string) {
	requestErrors := ctx.Value(ErrorsContextKey).(*ResponseErrors)
	requestErrors.Add(error)
}

func AddAndGetErrors(ctx context.Context, error string) *[]string {
	requestErrors := ctx.Value(ErrorsContextKey).(*ResponseErrors)

	if strings.Contains(error, "\n") {
		requestErrors.Errors = append(requestErrors.Errors, strings.Split(error, "\n")...)
	} else {
		requestErrors.Add(error)
	}

	return &requestErrors.Errors
}

func From(msg string) error {
	return errors.New(msg)
}
