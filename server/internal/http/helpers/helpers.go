package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"candly/internal/auth"
)

type HTTPMessage struct {
	Message string
}

func AppendJSONMessage(message string, extra map[string]interface{}) map[string]interface{} {
	extra["message"] = message
	return extra
}

func JSONMessage(message string) map[string]interface{} {
	return map[string]interface{}{
		"message": message,
	}

}

//credits: https://github.com/go-playground/validator/issues/559

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email"
	}
	return fe.Error() // default error
}

type ApiError struct {
	Param   string	`json:"param"`
	Message string	`json:"message"`
}

func SerializeValidationErr(err error) []ApiError {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]ApiError, len(ve))
		for i, fe := range ve {
			out[i] = ApiError{fe.Field(), msgForTag(fe)}
		}
		return out
	}

	return []ApiError{}

}


type ValidationError struct{
	Message string	`json:"message"`
	Errors []ApiError `json:"errors"`
}

func GenerateValidationError(err error) map[string]interface{} {
	return map[string]interface{}{
		"message": "validation error",
		"errors":  SerializeValidationErr(err),
	}
}

func GetUserClaims(c *gin.Context) (*auth.JwtUserClaims, bool) {
	cl, _ := c.Get("claims")
	claims, ok := cl.(*auth.JwtUserClaims)
	return claims, ok
}

func GetNewUserClaims(c *gin.Context) (*auth.JwtNewUserClaims, bool) {
	cl, _ := c.Get("claims")
	claims, ok := cl.(*auth.JwtNewUserClaims)
	return claims, ok
}
