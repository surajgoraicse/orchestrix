package response

import "github.com/labstack/echo/v5"

type apiResponse struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Success bool   `json:"success,omitempty"`
	Errors  string `json:"errors,omitempty"`
}

func NewResponse(c *echo.Context, statusCode int, message string, data any, err error) error {
	res := &apiResponse{
		Message: message,
		Data:    data,
		Success: true,
	}

	if statusCode < 200 || statusCode >= 300 || err != nil {
		res.Success = false
		res.Errors = err.Error()
	}

	return c.JSON(statusCode, res)
}