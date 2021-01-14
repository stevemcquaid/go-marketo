package marketo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// LeadResult default result struct
type LeadResult struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Created   string `json:"createdAt"`
	Updated   string `json:"updatedAt"`
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
