package errors

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    ErrorCode   `json:"code"`    // 错误码
	Message string      `json:"message"` // 消息
	Data    interface{} `json:"data"`    // 数据
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(200, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	appErr := GetAppError(err)
	
	// 过滤敏感信息：不返回底层错误详情给客户端
	// 只返回用户友好的错误消息
	c.JSON(appErr.HTTPStatus(), Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}

// ErrorWithCode 错误响应（指定错误码）
func ErrorWithCode(c *gin.Context, code ErrorCode, message string) {
	appErr := New(code, message)
	c.JSON(appErr.HTTPStatus(), Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}

// BadRequest 快捷返回400错误
func BadRequest(c *gin.Context, message string) {
	ErrorWithCode(c, CodeBadRequest, message)
}

// InvalidParam 快捷返回参数无效错误
func InvalidParam(c *gin.Context, paramName string) {
	ErrorWithCode(c, CodeInvalidParam, fmt.Sprintf("参数 %s 无效", paramName))
}

// MissingParam 快捷返回缺少参数错误
func MissingParam(c *gin.Context, paramName string) {
	ErrorWithCode(c, CodeMissingParam, fmt.Sprintf("缺少必要参数: %s", paramName))
}

// NotFound 快捷返回404错误
func NotFound(c *gin.Context, resource string) {
	ErrorWithCode(c, CodeNotFound, fmt.Sprintf("%s不存在", resource))
}

// InternalError 快捷返回500错误
func InternalError(c *gin.Context, message string) {
	ErrorWithCode(c, CodeInternalError, message)
}

// PaginatedData 分页数据结构
type PaginatedData struct {
	List  interface{} `json:"list"`  // 数据列表
	Total int         `json:"total"` // 总数
	Page  int         `json:"page"`  // 当前页
	Limit int         `json:"limit"` // 每页数量
}

// SuccessPaginated 成功的分页响应
func SuccessPaginated(c *gin.Context, list interface{}, total, page, limit int) {
	Success(c, PaginatedData{
		List:  list,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}
