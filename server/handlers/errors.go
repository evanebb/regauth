package handlers

import (
	"github.com/evanebb/regauth/oas"
	"net/http"
)

// newErrorResponse is a small wrapper function to create an oas.ErrorStatusCode with less boilerplate.
func newErrorResponse(statusCode int, message string) *oas.ErrorStatusCode {
	return &oas.ErrorStatusCode{StatusCode: statusCode, Response: oas.Error{Message: message}}
}

func newInternalServerErrorResponse() *oas.ErrorStatusCode {
	return newErrorResponse(http.StatusInternalServerError, "internal server error")
}
