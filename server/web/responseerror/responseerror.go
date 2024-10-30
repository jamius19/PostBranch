package responseerror

import (
	"context"
	"errors"
	"strings"
)

const ErrorsContextKey = "requestErrors"

type RequestErrors struct {
	Errors []string
}

func NewRequestErrors() *RequestErrors {
	return &RequestErrors{}
}

func (re *RequestErrors) Add(error string) {
	re.Errors = append(re.Errors, error)
}

func AddError(ctx context.Context, error string) {
	requestErrors := ctx.Value(ErrorsContextKey).(*RequestErrors)
	requestErrors.Add(error)
}

func AddAndGetErrors(ctx context.Context, error string) *[]string {
	requestErrors := ctx.Value(ErrorsContextKey).(*RequestErrors)

	if strings.Contains(error, "\n") {
		requestErrors.Errors = append(requestErrors.Errors, strings.Split(error, "\n")...)
	} else {
		requestErrors.Add(error)
	}

	return &requestErrors.Errors
}

func Clarify(msg string) error {
	return errors.New(msg)
}
