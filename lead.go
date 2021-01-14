package marketo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// LeadResult default result struct
type LeadResult struct {
	ID        int    `json:"id" mapstructure:"id"`
	FirstName string `json:"firstName" mapstructure:"firstName"`
	LastName  string `json:"lastName" mapstructure:"lastName"`
	Email     string `json:"email" mapstructure:"email"`
	Created   string `json:"createdAt" mapstructure:"createdAt"`
	Updated   string `json:"updatedAt" mapstructure:"updatedAt"`

	Fields map[string]string `json:"-" mapstructure:",remain"`
}

// LeadAttributeMap defines the name & readonly state of a Lead Attribute
type LeadAttributeMap struct {
	Name     string `json:"name"`
	ReadOnly bool   `json:"readOnly"`
}

// LeadAttribute is returned by the Describe Leads endpoint
type LeadAttribute struct {
	DataType    string           `json:"dataType"`
	DisplayName string           `json:"displayName"`
	ID          int              `json:"id"`
	Length      int              `json:"length"`
	REST        LeadAttributeMap `json:"rest"`
	SOAP        LeadAttributeMap `json:"soap"`
}

// LeadAttribute2 defines a lead attribute defined by the describe2.json
// endpoint.
type LeadAttribute2 struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	DataType    string `json:"dataType,omitempty"`
	Length      int    `json:"length,omitempty"`
	Updateable  bool   `json:"updateable,omitempty"`
	CRMManaged  bool   `json:"crmManaged,omitempty"`

	Searchable bool `json:"searchable,omitempty"`
}

// leadDescribe2Response contains the envelope used to deserialize a call to
// describe2.json
type leadDescribe2Response struct {
	Name             string
	SearchableFields [][]string
	Fields           []LeadAttribute2
}

const (
	describeLead2 = "describe2 lead"
	filterLeads   = "filter leads"
)

// LeadAPI provides access to the Marketo Lead API
type LeadAPI struct {
	c *Client
}

// NewLeadAPI returns a new instance of the lead API, configured with the
// provided Client.
func NewLeadAPI(c *Client) *LeadAPI {
	return &LeadAPI{c: c}
}

// DescribeFields fetches the Lead schema from Marketo and returns the set of
// attributes defined
func (l *LeadAPI) DescribeFields(ctx context.Context) ([]LeadAttribute2, error) {
	request, err := http.NewRequest(
		http.MethodGet, l.c.url("rest", "v1", "leads", "describe2.json"), nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := l.c.doRequest(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, handleError(describeLead2, resp)
	}

	response := &Response{}
	reader := json.NewDecoder(resp.Body)
	err = reader.Decode(response)
	if err != nil {
		return nil, err
	}

	object := []leadDescribe2Response{}
	err = json.Unmarshal(response.Result, &object)
	if len(object) == 0 {
		return nil, errors.New("not found")
	}

	searchable := map[string]bool{}
	for _, s := range object[0].SearchableFields {
		for _, fld := range s {
			searchable[fld] = true
		}
	}

	for i, field := range object[0].Fields {
		field.Searchable = searchable[field.Name]
		object[0].Fields[i] = field
	}
	return object[0].Fields, err
}

// Filter queries Marketo for one or more Leads, returning them if present
func (l *LeadAPI) Filter(ctx context.Context, opts ...QueryOption) ([]LeadResult, string, error) {
	q := &Query{}
	for _, opt := range opts {
		opt(q)
	}

	query, err := q.Values()
	if err != nil {
		return nil, "", err
	}
	request, err := http.NewRequest(
		http.MethodPost,
		l.c.url("rest", "v1", "leads.json?_method=GET"),
		strings.NewReader(query.Encode()),
	)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return nil, "", err
	}

	resp, err := l.c.doRequest(request)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", handleError(filterLeads, resp)
	}

	response := &Response{}
	reader := json.NewDecoder(resp.Body)
	err = reader.Decode(response)
	if err != nil {
		return nil, "", err
	}

	raw := []map[string]interface{}{}
	err = json.Unmarshal(response.Result, &raw)
	if err != nil {
		return nil, "", err
	}

	leads := make([]LeadResult, len(raw))
	for i, l := range raw {
		err = mapstructure.Decode(l, &leads[i])
		if err != nil {
			return nil, "", err
		}
	}

	return leads, response.NextPageToken, nil
}
