package marketo

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

const (
	// MaximumQueryBatchSize is the largest batch size requestable via Marketo's
	// list/query API.
	MaximumQueryBatchSize = 300
)

// Query contains the possible parameters used when listing Marketo objects
type Query struct {
	FilterField   string   `json:"filterType,omitempty"`
	FilterValues  []string `json:"filterValues,omitempty"`
	Fields        []string `json:"fields,omitempty"`
	BatchSize     int      `json:"batchSize,omitempty"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}

// Values returns the query payload as url.Values; if the query is invalid, an
// error is returned.
func (q *Query) Values() (url.Values, error) {
	result := url.Values{}
	if q.BatchSize == 0 {
		q.BatchSize = MaximumQueryBatchSize
	}

	if len(q.FilterValues) < 1 {
		return result, errors.New("too few values")
	}
	if len(q.FilterValues) > MaximumQueryBatchSize {
		return result, errors.New("too many values")
	}

	values := url.Values{}
	values.Set("filterType", q.FilterField)
	values.Set("filterValues", strings.Join(q.FilterValues, ","))
	if len(q.Fields) > 0 {
		values.Set("fields", strings.Join(q.Fields, ","))
	}
	values.Set("batchSize", strconv.Itoa(q.BatchSize))
	if q.NextPageToken != "" {
		values.Set("nextPageToken", q.NextPageToken)
	}

	return values, nil
}

// QueryOption defines the signature of functional options for Marketo Query
// APIs.
type QueryOption func(*Query)

// FilterField sets the field to search Marketo using
func FilterField(field string) QueryOption {
	return func(q *Query) {
		q.FilterField = field
	}
}

// FilterValues sets the possible values to match
func FilterValues(values []string) QueryOption {
	return func(q *Query) {
		q.FilterValues = values
	}
}

// GetFields sets the fields to retrieve for matching records
func GetFields(fields ...string) QueryOption {
	return func(q *Query) {
		q.Fields = fields
	}
}

// GetPage sets the paging token for the query
func GetPage(t string) QueryOption {
	return func(q *Query) {
		q.NextPageToken = t
	}
}
