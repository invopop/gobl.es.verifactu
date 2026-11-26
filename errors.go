package verifactu

import (
	"errors"
	"strings"
)

// Standard gateway error responses
var (
	// ErrConnection is used when the gateway could not be reached or gave a
	// response we're unable to understand.
	ErrConnection = newError("connection")
	// ErrValidation implies there is something wrong with the contents of the
	// request that needs to be fixed before sending it again.
	ErrValidation = newError("validation")
	// ErrDuplicate is used when the gateway has already received the request.
	ErrDuplicate = newError("duplicate")
	// ErrWarning means the request was accepted, but the gateway reported
	// issues that should be reviewed.
	ErrWarning = newError("warning")
	// ErrServer indicates the gateway had an internal problem handling the
	// request, which may succeed if attempted again later.
	ErrServer = newError("server")
)

// Standard error responses.
var (
	ErrNotSpanish       = ErrValidation.WithMessage("only spanish invoices are supported")
	ErrAlreadyProcessed = ErrValidation.WithMessage("already processed")
	ErrOnlyInvoices     = ErrValidation.WithMessage("only invoices are supported")
	ErrOnlyStatuses     = ErrValidation.WithMessage("only bill status documents are supported")
	ErrOnlyRecords      = ErrValidation.WithMessage("only invoice registration, invoice cancellation and event records are supported")
	ErrIncompleteRecord = ErrValidation.WithMessage("record is missing the details its fingerprint is computed from")
)

// Error allows for structured responses from the gateway to be able to
// response codes and messages.
type Error struct {
	key     string
	code    string
	message string
	cause   error
}

// Error produces a human readable error message.
func (e *Error) Error() string {
	out := []string{e.key}
	if e.code != "" {
		out = append(out, e.code)
	}
	if e.message != "" {
		out = append(out, e.message)
	}
	return strings.Join(out, ": ")
}

// Key returns the key for the error.
func (e *Error) Key() string {
	return e.key
}

// Message returns the human message for the error.
func (e *Error) Message() string {
	return e.message
}

// Code returns the code provided by the remote service.
func (e *Error) Code() string {
	return e.code
}

func newError(key string) *Error {
	return &Error{key: key}
}

// WithCode duplicates and adds the code to the error.
func (e *Error) WithCode(code string) *Error {
	e = e.clone()
	e.code = code
	return e
}

// WithMessage duplicates and adds the message to the error.
func (e *Error) WithMessage(msg string) *Error {
	e = e.clone()
	e.message = msg
	return e
}

func (e *Error) clone() *Error {
	ne := new(Error)
	*ne = *e
	return ne
}

// Is checks to see if the target error is the same as the current one
// or forms part of the chain.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return errors.Is(e.cause, target)
	}
	return e.key == t.key
}
