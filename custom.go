package marketo

import "time"

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
	Field string
	Name  string
}

type ObjectRelation struct {
	Field     string
	RelatedTo RelatedObject
	Type      string
}

type ObjectField struct {
	DataType    string
	DisplayName string
	Length      int
	Name        string
	Updateable  bool
	CRMManaged  bool
}

type CustomObjectMetadata struct {
	IDField          string
	APIName          string
	Description      string
	DisplayName      string
	PluralName       string
	Fields           []ObjectField
	SearchableFields []string
	DedupeFields     []string
	Relationships    []ObjectRelation
	CreatedAt        time.Time
	UpdatedAt        time.Time
	State            ObjectState
	Version          ObjectVersion
}
