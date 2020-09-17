package marketo

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
