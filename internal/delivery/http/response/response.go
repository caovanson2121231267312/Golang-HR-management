package response

import (
	"net/http"

	"hr-management-system/internal/i18n"
	"hr-management-system/internal/infrastructure/database"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string            `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

type Meta struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

func getLanguage(c *gin.Context) string {
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		lang = c.Query("lang")
	}
	if lang == "" {
		lang = "vi"
	}
	if lang != "vi" && lang != "en" {
		lang = "vi"
	}
	return lang
}

func OK(c *gin.Context, messageKey string, data interface{}) {
	lang := getLanguage(c)
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: i18n.T(lang, messageKey),
		Data:    data,
	})
}

func OKWithMeta(c *gin.Context, messageKey string, data interface{}, pagination *database.Pagination) {
	lang := getLanguage(c)
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: i18n.T(lang, messageKey),
		Data:    data,
		Meta: &Meta{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			Total:      pagination.Total,
			TotalPages: pagination.Pages,
		},
	})
}

func Created(c *gin.Context, messageKey string, data interface{}) {
	lang := getLanguage(c)
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: i18n.T(lang, messageKey),
		Data:    data,
	})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func BadRequest(c *gin.Context, messageKey string, details map[string]string) {
	lang := getLanguage(c)
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: i18n.T(lang, messageKey),
		Error:   &ErrorInfo{Code: "BAD_REQUEST", Details: details},
	})
}

func Unauthorized(c *gin.Context, messageKey string) {
	lang := getLanguage(c)
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Message: i18n.T(lang, messageKey),
		Error:   &ErrorInfo{Code: "UNAUTHORIZED"},
	})
}

func Forbidden(c *gin.Context, messageKey string) {
	lang := getLanguage(c)
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Message: i18n.T(lang, messageKey),
		Error:   &ErrorInfo{Code: "FORBIDDEN"},
	})
}

func NotFound(c *gin.Context, messageKey string) {
	lang := getLanguage(c)
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Message: i18n.T(lang, messageKey),
		Error:   &ErrorInfo{Code: "NOT_FOUND"},
	})
}

func Conflict(c *gin.Context, messageKey string) {
	lang := getLanguage(c)
	c.JSON(http.StatusConflict, Response{
		Success: false,
		Message: i18n.T(lang, messageKey),
		Error:   &ErrorInfo{Code: "CONFLICT"},
	})
}

func UnprocessableEntity(c *gin.Context, messageKey string, details map[string]string) {
	lang := getLanguage(c)
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Message: i18n.T(lang, messageKey),
		Error:   &ErrorInfo{Code: "VALIDATION_ERROR", Details: details},
	})
}

func TooManyRequests(c *gin.Context, messageKey string) {
	lang := getLanguage(c)
	c.JSON(http.StatusTooManyRequests, Response{
		Success: false,
		Message: i18n.T(lang, messageKey),
		Error:   &ErrorInfo{Code: "RATE_LIMIT_EXCEEDED"},
	})
}

func InternalError(c *gin.Context, err error) {
	lang := getLanguage(c)
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Message: i18n.T(lang, "common.internal_error"),
		Error:   &ErrorInfo{Code: "INTERNAL_ERROR"},
	})
}

func ServiceUnavailable(c *gin.Context) {
	lang := getLanguage(c)
	c.JSON(http.StatusServiceUnavailable, Response{
		Success: false,
		Message: i18n.T(lang, "common.error"),
		Error:   &ErrorInfo{Code: "SERVICE_UNAVAILABLE"},
	})
}

func ValidationError(c *gin.Context, errors map[string]string) {
	lang := getLanguage(c)
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Message: i18n.T(lang, "common.validation_error"),
		Error:   &ErrorInfo{Code: "VALIDATION_ERROR", Details: errors},
	})
}

func Error(c *gin.Context, statusCode int, code, messageKey string, details map[string]string) {
	lang := getLanguage(c)
	c.JSON(statusCode, Response{
		Success: false,
		Message: i18n.T(lang, messageKey),
		Error:   &ErrorInfo{Code: code, Details: details},
	})
}

func RawOK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func RawError(c *gin.Context, statusCode int, code, message string, details map[string]string) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   &ErrorInfo{Code: code, Details: details},
	})
}
