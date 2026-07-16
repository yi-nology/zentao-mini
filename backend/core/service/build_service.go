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
	const pageSize = 100
	var allBuilds []zentao.Build

	if query.ProjectID > 0 {
		allBuilds, _ = s.fetchBuildsByProject(query.ProjectID, pageSize)
	} else if query.ExecutionID > 0 {
		allBuilds, _ = s.fetchBuildsByExecution(query.ExecutionID, pageSize)
	} else {
		return []vo.BuildVO{}, nil
	}

	return s.convertToVO(allBuilds), nil
}

func (s *BuildService) fetchBuildsByProject(projectID, pageSize int) ([]zentao.Build, error) {
	var allBuilds []zentao.Build
	page := 1
	for {
		builds, err := s.client.GetBuildsByProject(projectID, page, pageSize)
		if err != nil {
			return nil, err
		}
		allBuilds = append(allBuilds, builds...)
		if len(builds) < pageSize {
			break
		}
		page++
	}
	return allBuilds, nil
}

func (s *BuildService) fetchBuildsByExecution(executionID, pageSize int) ([]zentao.Build, error) {
	var allBuilds []zentao.Build
	page := 1
	for {
		builds, err := s.client.GetBuildsByExecution(executionID, page, pageSize)
		if err != nil {
			return nil, err
		}
		allBuilds = append(allBuilds, builds...)
		if len(builds) < pageSize {
			break
		}
		page++
	}
	return allBuilds, nil
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
