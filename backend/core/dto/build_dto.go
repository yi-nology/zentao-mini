package dto

type BuildQueryDTO struct {
	ProjectID   int `form:"projectId" json:"projectId"`
	ExecutionID int `form:"executionId" json:"executionId"`
}

func (dto *BuildQueryDTO) Validate() error {
	return nil
}
