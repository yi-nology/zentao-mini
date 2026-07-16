package zentao

import (
	"github.com/yi-nology/common/biz/zentao"
)

// GetBuildsByProject 获取项目下的版本列表
func (c *Client) GetBuildsByProject(projectID int, page, pageSize int) ([]zentao.Build, error) {
	var response *zentao.BuildListResponse
	err := c.withTokenRetry("GetBuildsByProject", func(client *zentao.Client) error {
		var err error
		response, err = client.GetBuildsByProject(projectID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Builds, nil
}

// GetBuildsByExecution 获取执行下的版本列表
func (c *Client) GetBuildsByExecution(executionID int, page, pageSize int) ([]zentao.Build, error) {
	var response *zentao.BuildListResponse
	err := c.withTokenRetry("GetBuildsByExecution", func(client *zentao.Client) error {
		var err error
		response, err = client.GetBuildsByExecution(executionID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Builds, nil
}
