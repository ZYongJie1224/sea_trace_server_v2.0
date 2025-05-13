package utils

// Response 标准API响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(message string) *Response {
	return &Response{
		Code:    500,
		Message: message,
	}
}

// UnauthorizedResponse 未授权响应
func UnauthorizedResponse() *Response {
	return &Response{
		Code:    401,
		Message: "未授权的访问",
	}
}

// ForbiddenResponse 禁止访问响应
func ForbiddenResponse() *Response {
	return &Response{
		Code:    403,
		Message: "权限不足",
	}
}
