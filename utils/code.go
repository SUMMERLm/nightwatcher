package utils

import (
	"net/http"

	"github.com/marmotedu/errors"
	"github.com/novalagung/gubrak"
)

// TODO: go generate

// Common: basic errors.
// Code must start with 10.
const (
	// ErrSuccess - 200: OK.
	ErrSuccess int = iota + 100001

	// ErrUnknown - 500: Internal server error.
	ErrUnknown

	// ErrBind - 400: RpcError occurred while binding the request body to the struct.
	ErrBind

	// ErrValidation - 400: Validation failed.
	ErrValidation
)

// common: database errors.
const (
	// ErrDatabase - 500: Database error.
	ErrDatabase int = iota + 100101
)

// common: network errors.
const (
	// ErrNetwork - 500: Network error.
	ErrNetwork int = iota + 100201
)

// common: encode/decode errors.
const (
	// ErrEncodingFailed - 500: Encoding failed due to an error with the data.
	ErrEncodingFailed int = iota + 100301

	// ErrDecodingFailed - 500: Decoding failed due to an error with the data.
	ErrDecodingFailed

	// ErrInvalidJSON - 500: Data is not valid JSON.
	ErrInvalidJSON

	// ErrEncodingJSON - 500: JSON data could not be encoded.
	ErrEncodingJSON

	// ErrDecodingJSON - 500: JSON data could not be decoded.
	ErrDecodingJSON

	// ErrInvalidYaml - 500: Data is not valid Yaml.
	ErrInvalidYaml

	// ErrEncodingYaml - 500: Yaml data could not be encoded.
	ErrEncodingYaml

	// ErrDecodingYaml - 500: Yaml data could not be decoded.
	ErrDecodingYaml
)

// common: authorization and authentication errors.
const (
	// ErrTokenInvalid - 401: Token invalid.
	ErrTokenInvalid int = iota + 100401

	// ErrEncrypt - 401: Error occurred while encrypting the user password.
	ErrEncrypt

	// ErrSignatureInvalid - 401: Signature is invalid.
	ErrSignatureInvalid

	// ErrExpired - 401: Token expired.
	ErrExpired

	// ErrInvalidAuthHeader - 401: Invalid authorization header.
	ErrInvalidAuthHeader

	// ErrMissingHeader - 401: The `Authorization` header was empty.
	ErrMissingHeader

	// ErrPasswordIncorrect - 401: Password was incorrect.
	ErrPasswordIncorrect

	// 	ErrPermissionDenied - 403: Permission denied.
	ErrPermissionDenied

	// ErrInBlack - 401: User is in black list.
	ErrInBlack

	// ErrHasLogin - 401: User has login
	ErrHasLogin

	// ErrSendSmsCode - 401: Send sms code failed.
	ErrSendSmsCode

	// ErrInvalidSmsCode - 401: Sms code is invalid.
	ErrInvalidSmsCode
)

// smilelink errors.
// All smilelink-specific error codes start with 12
const (
	// ErrObs - 500: Call obs service failed.
	ErrObs = iota + 120001

	// ErrChatLimit - 400: User send message exceed.
	ErrChatLimit
)

// ErrCode implements `github.com/marmotedu/errors`.Coder interface.
type ErrCode struct {
	// C refers to the code of the ErrCode.
	C int

	// HTTP status that should be used for the associated error code.
	HTTP int

	// External (user) facing error text.
	Ext string

	// Ref specify the reference document.
	Ref string
}

var _ errors.Coder = &ErrCode{}

// Code returns the integer code of ErrCode.
func (coder ErrCode) Code() int {
	return coder.C
}

// String implements stringer. String returns the external error message,
// if any.
func (coder ErrCode) String() string {
	return coder.Ext
}

// Reference returns the reference document.
func (coder ErrCode) Reference() string {
	return coder.Ref
}

// HTTPStatus returns the associated HTTP status code, if any. Otherwise,
// returns 200.
func (coder ErrCode) HTTPStatus() int {
	if coder.HTTP == 0 {
		return http.StatusInternalServerError
	}

	return coder.HTTP
}

// nolint: unparam
func register(code int, httpStatus int, message string, refs ...string) {
	found, _ := gubrak.Includes([]int{200, 400, 401, 403, 404, 500}, httpStatus)
	if !found {
		panic("http code not in `200, 400, 401, 403, 404, 500`")
	}

	var reference string
	if len(refs) > 0 {
		reference = refs[0]
	}

	coder := &ErrCode{
		C:    code,
		HTTP: httpStatus,
		Ext:  message,
		Ref:  reference,
	}

	errors.MustRegister(coder)
}

type RpcError struct {
	C    int `json:"code,omitempty"`
	HTTP int
	Ext  string `json:"message,omitempty"`
	Ref  string `json:"reference,omitempty"`
}

func (e RpcError) HTTPStatus() int {
	return e.HTTP
}

func (e RpcError) String() string {
	return e.Ext
}

func (e RpcError) Reference() string {
	return e.Ref
}

func (e RpcError) Code() int {
	return e.C
}

func (e RpcError) Error() string {
	return e.Ext
}

func init() {
	register(ErrSuccess, 200, "成功")
	register(ErrUnknown, 500, "服务器错误")
	register(ErrBind, 400, "请求参数有误")
	register(ErrValidation, 400, "参数验证失败")
	register(ErrDatabase, 500, "数据库错误")
	register(ErrNetwork, 500, "网络错误")
	register(ErrEncodingJSON, 500, "编码 JSON 数据出错")
	register(ErrDecodingJSON, 500, "解码 JSON 数据出错")
	register(ErrTokenInvalid, 401, "Token 不合法")
	register(ErrExpired, 401, "Token 已过期")
	register(ErrSignatureInvalid, 401, "签名不合法")
	register(ErrMissingHeader, 401, "缺少认证头部")
	register(ErrPasswordIncorrect, 401, "密码不正确")
	register(ErrPermissionDenied, 403, "没有操作权限")
	register(ErrInBlack, 401, "用户已被封禁")
	register(ErrHasLogin, 401, "用户已在其他地方登录")
	register(ErrSendSmsCode, 500, "发送短信验证码失败")
	register(ErrInvalidSmsCode, 400, "短信验证码错误")

	register(ErrObs, 500, "调用对象存储服务出错")
	register(ErrChatLimit, 400, "用户当日发送消息已达系统上限")
}
