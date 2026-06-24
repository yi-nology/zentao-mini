package dto

import "fmt"

type TaskQueryDTO struct {
	ProductID   int    `form:"productId" json:"productId"`
	ExecutionID int    `form:"executionId" json:"executionId"`
	AssignedTo  string `form:"assignedTo" json:"assignedTo"`
	Status      string `form:"status" json:"status"`
	StartDate   string `form:"startDate" json:"startDate"`
	EndDate     string `form:"endDate" json:"endDate"`
	Page        int    `form:"page" json:"page"`
	PageSize    int    `form:"pageSize" json:"pageSize"`
}

func (dto *TaskQueryDTO) Validate() error {
	if dto.Page <= 0 {
		dto.Page = 1
	}
	if dto.PageSize <= 0 {
		dto.PageSize = 20
	}
	if dto.PageSize > MaxPageSize {
		dto.PageSize = MaxPageSize
	}
	if dto.ProductID <= 0 {
		return fmt.Errorf("需要选择产品")
	}
	return nil
}
