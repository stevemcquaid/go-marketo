package marketo

import (
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

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

	return Error{
		Message:    "error creating import",
		Body:       string(body),
		StatusCode: resp.StatusCode,
	}
}
