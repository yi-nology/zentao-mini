package service

import (
	"github.com/yi-nology/common/biz/zentao"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/vo"
	myzentao "github.com/yi-nology/zentao-mini/backend/core/zentao"
)

type BuildService struct {
	client *myzentao.Client
}

func NewBuildService(client *myzentao.Client) *BuildService {
	return &BuildService{client: client}
}

func (s *BuildService) GetBuilds(query *dto.BuildQueryDTO) ([]vo.BuildVO, error) {
	var builds []zentao.Build
	var err error

	if query.ProjectID > 0 {
		builds, err = s.client.GetBuildsByProject(query.ProjectID, 1, 1000)
	} else if query.ExecutionID > 0 {
		builds, err = s.client.GetBuildsByExecution(query.ExecutionID, 1, 1000)
	} else {
		return []vo.BuildVO{}, nil
	}

	if err != nil {
		return nil, err
	}

	return s.convertToVO(builds), nil
}

func (s *BuildService) convertToVO(builds []zentao.Build) []vo.BuildVO {
	if len(builds) == 0 {
		return []vo.BuildVO{}
	}

	result := make([]vo.BuildVO, 0, len(builds))
	for _, build := range builds {
		result = append(result, vo.BuildVO{
			ID:      build.ID,
			Product: build.Product,
			Project: build.Project,
			Name:    build.Name,
			Date:    build.Date,
		})
	}
	return result
}
