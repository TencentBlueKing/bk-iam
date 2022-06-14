package types

type ResourceChangedContent struct {
	ResourceTypePK int64
	ResourceID     string

	ActionRelatedResourceTypePK int64
	CreatedActionPKs            []int64
	DeletedActionPKs            []int64
}
