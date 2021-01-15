package marketo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// A Reason is provided along with errors, warnings, and some operations
type Reason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r Reason) Error() string {
	return fmt.Sprintf("%s: %s", r.Code, r.Message)
}

var (
	ErrBadGateway                    = Reason{Code: "502"}
	ErrEmptyAccessToken              = Reason{Code: "600"}
	ErrAccessTokenInvalid            = Reason{Code: "601"}
	ErrAccessTokenExpired            = Reason{Code: "602"}
	ErrAccessDenied                  = Reason{Code: "603"}
	ErrRequestTimeOut                = Reason{Code: "604"}
	ErrMethodUnsupported             = Reason{Code: "605"}
	ErrRateLimitExceeded             = Reason{Code: "606"}
	ErrDailyQuotaReached             = Reason{Code: "607"}
	ErrTemporarilyUnavailable        = Reason{Code: "608"}
	ErrInvalidJSON                   = Reason{Code: "609"}
	ErrNotFound                      = Reason{Code: "610"}
	ErrSystemError                   = Reason{Code: "611"}
	ErrInvalidContentType            = Reason{Code: "612"}
	ErrInvalidMultipart              = Reason{Code: "613"}
	ErrInvalidSubscription           = Reason{Code: "614"}
	ErrConcurrentLimitReached        = Reason{Code: "615"}
	ErrInvalidSubscriptionType       = Reason{Code: "616"}
	ErrCannotBeBlank                 = Reason{Code: "701"}
	ErrNoDataFound                   = Reason{Code: "702"}
	ErrFeatureNotEnabled             = Reason{Code: "703"}
	ErrInvalidDateFormat             = Reason{Code: "704"}
	ErrBusinessRuleViolation         = Reason{Code: "709"}
	ErrParentFolderNotFound          = Reason{Code: "710"}
	ErrIncompatibleFolderType        = Reason{Code: "711"}
	ErrMergeOperationInvalid         = Reason{Code: "712"}
	ErrTransientError                = Reason{Code: "713"}
	ErrUnableToFindDefaultRecordType = Reason{Code: "714"}
	ErrExternalSalesPersonIDNotFound = Reason{Code: "718"}

	ErrTooManyImports = Reason{Code: "1016"}
)

// Error contains the error state returned from a Marketo operation
type Error struct {
	Message    string
	StatusCode int
	Body       string

	Errors []Reason
}

// ErrorForReasons returns a new Error wrapping the Reasons provided by the
// Marketo API
func ErrorForReasons(status int, reasons ...Reason) Error {
	return Error{
		StatusCode: status,
		Errors:     reasons,
	}
}

// Is provides support for the errors.Is() call, and will return true if the
// passed target is a Reason and it matches any of the Reasons included with
// this Error.
func (e Error) Is(target error) bool {
	if reason, ok := target.(Reason); ok {
		for _, r := range e.Errors {
			if r.Code == reason.Code {
				return true
			}
		}
	}
	return false
}

// Error fulfills the error interface
func (e Error) Error() string {
	if e.Message != "" {
		return e.Message
	}

	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Message
	}
	return strings.Join(msgs, "; ")
}

// handleError reads a non-successful HTTP responnse & returns an
// error wrapping it; it is the callers responsibility to close
// response.Body.
func handleError(operation string, resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "unable to read marketo error response")
	}

	// attempt to deserialize the error response
	response := Response{}
	err = json.Unmarshal(body, &response)
	if err == nil {
		return ErrorForReasons(resp.StatusCode, response.Errors...)
	}

	return Error{
		Message:    fmt.Sprintf("error: %s", operation),
		Body:       string(body),
		StatusCode: resp.StatusCode,
	}
}
