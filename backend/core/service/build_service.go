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
	var err error

	if query.ProjectID > 0 {
		allBuilds, err = s.fetchAllBuilds(pageSize, func(page, size int) ([]zentao.Build, error) {
			return s.client.GetBuildsByProject(query.ProjectID, page, size)
		})
	} else if query.ExecutionID > 0 {
		allBuilds, err = s.fetchAllBuilds(pageSize, func(page, size int) ([]zentao.Build, error) {
			return s.client.GetBuildsByExecution(query.ExecutionID, page, size)
		})
	} else {
		return []vo.BuildVO{}, nil
	}

	if err != nil {
		return nil, err
	}

	return s.convertToVO(allBuilds), nil
}

// fetchAllBuilds 通用分页查询函数，复用于不同查询条件的版本列表获取
func (s *BuildService) fetchAllBuilds(pageSize int, fetchFn func(page, pageSize int) ([]zentao.Build, error)) ([]zentao.Build, error) {
	var allBuilds []zentao.Build
	page := 1
	for {
		builds, err := fetchFn(page, pageSize)
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
