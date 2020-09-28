package marketo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type ObjectState string

const (
	ObjectStateDraft             ObjectState = "draft"
	ObjectStateApproved          ObjectState = "approved"
	ObjectStateApprovedWithDraft ObjectState = "approvedWithDraft"
)

type ObjectVersion string

const (
	DraftVersion    ObjectVersion = "draft"
	ApprovedVersion ObjectVersion = "approved"
)

type RelatedObject struct {
	Field string `json:"field"`
	Name  string `json:"name"`
}

type ObjectRelation struct {
	Field     string        `json:"field"`
	RelatedTo RelatedObject `json:"relatedTo"`
	Type      string        `json:"type"`
}

type ObjectField struct {
	DataType    string `json:"dataType"`
	DisplayName string `json:"displayName"`
	Length      int    `json:"length"`
	Name        string `json:"name"`
	Updateable  bool   `json:"updateable"`
	CRMManaged  bool   `json:"crmManaged"`
}

type CustomObjectMetadata struct {
	IDField          string           `json:"idField"`
	APIName          string           `json:"name"`
	Description      string           `json:"description"`
	DisplayName      string           `json:"displayName"`
	PluralName       string           `json:"pluralName"`
	Fields           []ObjectField    `json:"fields"`
	SearchableFields [][]string       `json:"searchableFields"`
	DedupeFields     []string         `json:"dedupeFields"`
	Relationships    []ObjectRelation `json:"relationships"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
	State            ObjectState      `json:"state"`
	Version          ObjectVersion    `json:"version"`
}

// CustomObjects provides access to the Marketo custom objects API
type CustomObjects struct {
	*Client
}

// NewCustomObjectsAPI returns a new instance of the
func NewCustomObjectsAPI(c *Client) *CustomObjects {
	return &CustomObjects{c}
}

func (c *CustomObjects) List(ctx context.Context) ([]CustomObjectMetadata, error) {
	request, err := http.NewRequest(
		http.MethodGet, c.url("rest", "v1", "customobjects.json"), nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.doRequest(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, handleError(getLeadImport, resp)
	}

	response := &Response{}
	reader := json.NewDecoder(resp.Body)
	err = reader.Decode(response)
	if err != nil {
		return nil, err
	}

	objects := []CustomObjectMetadata{}
	err = json.Unmarshal(response.Result, &objects)

	return objects, err

}

func (c *CustomObjects) Describe(ctx context.Context, name string) (*CustomObjectMetadata, error) {
	request, err := http.NewRequest(
		http.MethodGet, c.url("rest", "v1", "customobjects", name, "describe.json"), nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.doRequest(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, handleError(getLeadImport, resp)
	}

	response := &Response{}
	reader := json.NewDecoder(resp.Body)
	err = reader.Decode(response)
	if err != nil {
		return nil, err
	}

	object := []CustomObjectMetadata{}
	err = json.Unmarshal(response.Result, &object)

	if len(object) == 0 {
		return nil, errors.New("not found")
	}
	return &object[0], err

}
