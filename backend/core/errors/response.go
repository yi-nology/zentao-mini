package errors

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(200, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, err error) {
	appErr := GetAppError(err)

	c.JSON(appErr.HTTPStatus(), Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}

func ErrorWithCode(c *gin.Context, code ErrorCode, message string) {
	appErr := New(code, message)
	c.JSON(appErr.HTTPStatus(), Response{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}

func BadRequest(c *gin.Context, message string) {
	ErrorWithCode(c, CodeBadRequest, message)
}

func InvalidParam(c *gin.Context, paramName string) {
	ErrorWithCode(c, CodeInvalidParam, fmt.Sprintf("参数 %s 无效", paramName))
}

func MissingParam(c *gin.Context, paramName string) {
	ErrorWithCode(c, CodeMissingParam, fmt.Sprintf("缺少必要参数: %s", paramName))
}

func NotFound(c *gin.Context, resource string) {
	ErrorWithCode(c, CodeNotFound, fmt.Sprintf("%s不存在", resource))
}

func InternalError(c *gin.Context, message string) {
	ErrorWithCode(c, CodeInternalError, message)
}

type PaginatedData struct {
	List     interface{} `json:"list"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

func SuccessPaginated(c *gin.Context, list interface{}, total, page, pageSize int) {
	Success(c, PaginatedData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
