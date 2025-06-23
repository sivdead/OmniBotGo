package v1

import "net/http"

const (
	// HTTP状态码常量 - 成功状态
	StatusOK       = http.StatusOK       // 200
	StatusCreated  = http.StatusCreated  // 201
	StatusAccepted = http.StatusAccepted // 202

	// HTTP状态码常量 - 客户端错误
	StatusBadRequest          = http.StatusBadRequest          // 400
	StatusUnauthorized        = http.StatusUnauthorized        // 401
	StatusForbidden           = http.StatusForbidden           // 403
	StatusNotFound            = http.StatusNotFound            // 404
	StatusMethodNotAllowed    = http.StatusMethodNotAllowed    // 405
	StatusConflict            = http.StatusConflict            // 409
	StatusTooManyRequests     = http.StatusTooManyRequests     // 429
	StatusUnprocessableEntity = http.StatusUnprocessableEntity // 422

	// HTTP状态码常量 - 服务器错误
	StatusInternalServerError = http.StatusInternalServerError // 500
	StatusBadGateway          = http.StatusBadGateway          // 502
	StatusServiceUnavailable  = http.StatusServiceUnavailable  // 503
	StatusGatewayTimeout      = http.StatusGatewayTimeout      // 504
)

const (
	// 业务错误代码
	CodeSuccess           = 0     // 成功
	CodeInvalidParams     = 40001 // 参数错误
	CodeNotFound          = 40401 // 资源不存在
	CodeInternalError     = 50001 // 内部错误
	CodeDatabaseError     = 50002 // 数据库错误
	CodeExternalAPIError  = 50003 // 外部API错误
	CodeValidationError   = 40002 // 验证错误
	CodeAuthenticationErr = 40101 // 认证错误
	CodeAuthorizationErr  = 40301 // 授权错误
)

const (
	// 常用错误消息
	MsgSuccess           = "操作成功"
	MsgInvalidParams     = "参数错误"
	MsgNotFound          = "资源不存在"
	MsgInternalError     = "内部服务器错误"
	MsgDatabaseError     = "数据库操作失败"
	MsgExternalAPIError  = "外部API调用失败"
	MsgValidationError   = "数据验证失败"
	MsgAuthenticationErr = "认证失败"
	MsgAuthorizationErr  = "权限不足"
	MsgTooManyRequests   = "请求过于频繁，请稍后再试"
)