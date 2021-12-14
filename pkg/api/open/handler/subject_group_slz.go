package handler

type subjectGroupsSerializer struct {
	// NOTE: currently only support subject_type=user, maybe support subject_type=department later
	SubjectType string `uri:"subject_type" binding:"required,oneof=user" example:"user"`
	SubjectID   string `uri:"subject_id" binding:"required" example:"admin"`
}

type responseSubject struct {
	Type string `json:"type" example:"user"`
	ID   string `json:"id" example:"admin"`
	Name string `json:"name" example:"Administer"`
}

type subjectGroupsResponse []responseSubject
