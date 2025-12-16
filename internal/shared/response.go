package shared

// GenerateResponse 统一响应包装
type GenerateResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 标准成功响应
func SuccessResponse(data interface{}) GenerateResponse {
	return GenerateResponse{
		Code: 0,
		Data: data,
	}
}

// ErrorResponse 标准错误响应
func ErrorResponse(code int, message string) GenerateResponse {
	return GenerateResponse{
		Code:    code,
		Message: message,
	}
}
