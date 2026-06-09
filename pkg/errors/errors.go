package errors

import (
	"fmt"
	"net/http"
)

type Problem struct {
	Type     string            `json:"type"`
	Title    string            `json:"title"`
	Status   int               `json:"status"`
	Detail   string            `json:"detail"`
	Instance string            `json:"instance,omitempty"`
	Code     string            `json:"code,omitempty"`
	Errors   map[string]string `json:"errors,omitempty"`
}

func (p *Problem) Error() string {
	return fmt.Sprintf("[%s] %s: %s", p.Code, p.Title, p.Detail)
}

var (
	ErrAccountNotFound      = NewNotFound("ACCOUNT_NOT_FOUND", "账户不存在")
	ErrAccountFrozen        = NewForbidden("ACCOUNT_FROZEN", "账户已冻结")
	ErrAccountClosed        = NewForbidden("ACCOUNT_CLOSED", "账户已销户")
	ErrInsufficientBalance  = NewBadRequest("INSUFFICIENT_BALANCE", "余额不足")
	ErrInvalidPassword      = NewBadRequest("INVALID_PASSWORD", "密码错误")
	ErrDuplicateBizID       = NewBadRequest("DUPLICATE_BIZ_ID", "重复的业务编号")
	ErrRiskLimitExceeded    = NewForbidden("RISK_LIMIT_EXCEEDED", "超出交易限额")
	ErrBlacklisted          = NewForbidden("BLACKLISTED", "账户在黑名单中")
	ErrSuspiciousTransaction = NewForbidden("SUSPICIOUS_TRANSACTION", "可疑交易")
	ErrInvalidAmount        = NewBadRequest("INVALID_AMOUNT", "金额无效")
	ErrSameAccount          = NewBadRequest("SAME_ACCOUNT", "不能给自己转账")
	ErrIdempotentConflict   = NewConflict("IDEMPOTENT_CONFLICT", "幂等冲突")
)

func NewBadRequest(code, detail string) *Problem {
	return &Problem{
		Type:   "https://tools.ietf.org/html/rfc7807",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Code:   code,
		Detail: detail,
	}
}

func NewNotFound(code, detail string) *Problem {
	return &Problem{
		Type:   "https://tools.ietf.org/html/rfc7807",
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Code:   code,
		Detail: detail,
	}
}

func NewForbidden(code, detail string) *Problem {
	return &Problem{
		Type:   "https://tools.ietf.org/html/rfc7807",
		Title:  "Forbidden",
		Status: http.StatusForbidden,
		Code:   code,
		Detail: detail,
	}
}

func NewConflict(code, detail string) *Problem {
	return &Problem{
		Type:   "https://tools.ietf.org/html/rfc7807",
		Title:  "Conflict",
		Status: http.StatusConflict,
		Code:   code,
		Detail: detail,
	}
}

func NewInternal(detail string) *Problem {
	return &Problem{
		Type:   "https://tools.ietf.org/html/rfc7807",
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
		Code:   "INTERNAL_ERROR",
		Detail: detail,
	}
}

func Wrap(err error, detail string) *Problem {
	if p, ok := err.(*Problem); ok {
		return p
	}
	return NewInternal(detail + ": " + err.Error())
}

func As(err error, target **Problem) bool {
	if p, ok := err.(*Problem); ok {
		*target = p
		return true
	}
	return false
}
