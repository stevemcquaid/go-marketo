package marketo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// ErrCode is the numeric error code returned by Marketo
type ErrCode int

const (
	ErrCodeBadGateway                    = 502
	ErrCodeEmptyAccessToken              = 600
	ErrCodeAccessTokenInvalid            = 601
	ErrCodeAccessTokenExpired            = 602
	ErrCodeAccessDenied                  = 603
	ErrCodeRequestTimeOut                = 604
	ErrCodeMethodUnsupported             = 605
	ErrCodeRateLimitExceeded             = 606
	ErrCodeDailyQuotaReached             = 607
	ErrCodeTemporarilyUnavailable        = 608
	ErrCodeInvalidJSON                   = 609
	ErrCodeNotFound                      = 610
	ErrCodeSystemError                   = 611
	ErrCodeInvalidContentType            = 612
	ErrCodeInvalidMultipart              = 613
	ErrCodeInvalidSubscription           = 614
	ErrCodeConcurrentLimitReached        = 615
	ErrCodeInvalidSubscriptionType       = 616
	ErrCodeCannotBeBlank                 = 701
	ErrCodeNoDataFound                   = 702
	ErrCodeFeatureNotEnabled             = 703
	ErrCodeInvalidDateFormat             = 704
	ErrCodeBusinessRuleViolation         = 709
	ErrCodeParentFolderNotFound          = 710
	ErrCodeIncompatibleFolderType        = 711
	ErrCodeMergeOperationInvalid         = 712
	ErrCodeTransientError                = 713
	ErrCodeUnableToFindDefaultRecordType = 714
	ErrCodeExternalSalesPersonIDNotFound = 718
)

// ErrorResponse contains the payload of a Marketo error response. The
// response itself may have 1 or more ResponseErrors with detailed
// error information.
//
// See https://developers.marketo.com/rest-api/error-codes/ for
// details on Marketo error types.
type ErrorResponse struct {
	RequestID string          `json:"requestId"`
	Success   bool            `json:"success"`
	Errors    []ResponseError `json:"errors,omitempty"`
}

// Error fulfills the error interface
func (e ErrorResponse) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Message
	}
	return strings.Join(msgs, "; ")
}

// ResponseError holds a single response-level error
type ResponseError struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
}

// Error contains the error state returned from a Marketo operation
type Error struct {
	Message    string
	StatusCode int
	Body       string
}

// Error fulfills the error interface
func (e Error) Error() string {
	return e.Message
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
	mktoErr := ErrorResponse{}
	err = json.Unmarshal(body, &mktoErr)
	if err == nil {
		return mktoErr
	}

	return Error{
		Message:    fmt.Sprintf("error: %s", operation),
		Body:       string(body),
		StatusCode: resp.StatusCode,
	}
}
