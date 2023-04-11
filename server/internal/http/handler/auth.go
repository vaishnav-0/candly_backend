package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"candly/internal/auth"
	"candly/internal/http/helpers"
)

type GenerateOTPBody struct {
	Phone string `json:"phone"`
}

type VerifyOTPBody struct {
	Phone string `json:"phone"`
	Otp   string `json:"otp"`
}

func (h *Handlers) GenerateOTP(c *gin.Context) {

	body := GenerateOTPBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, helpers.JSONMessage("mobile number is required"))
		return
	}

	otp, err := h.auth.GenerateOTP()

	if err != nil {
		h.log.Error().Err(err).Msg("cannot generate OPT")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = h.auth.StoreOTP(body.Phone, otp)

	if err == auth.ErrOTPLimit {
		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("OTP limit exceeded"))
		return
	} else if err == auth.ErrOTPRetry {
		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("OTP wait time not reached"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"otp": otp})

}

func (h *Handlers) VerifyOTP(c *gin.Context) {

	body := VerifyOTPBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, helpers.JSONMessage("mobile number and otp is required"))
		return
	}

	if err := h.auth.VerifyOTP(body.Phone, body.Otp); err == nil {

		token, err := h.auth.GenerateJWT(body.Phone)

		if err != nil {
			h.log.Error().Err(err).Msg("cannot generate jwt")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, gin.H{"access_token": token})

	} else if err == auth.ErrOTPInvalid {

		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("invalid otp"))

	} else if err == auth.ErrOTPRetry {

		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("otp retries exceeded"))

	}

}

type RegisterUserBody struct {
	Name  string `json:"name"  binding:"required"`
	Email string `json:"email" binding:"email"`
}

func (h *Handlers) RegisterUser(c *gin.Context) {
	body := RegisterUserBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, helpers.GenerateValidationError(err))
		return
	}
	c.Get("claims")

	claims, ok := helpers.GetNewUserClaims(c)

	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.JSONMessage("cannot parse token"))
		return
	}

	if err := h.auth.RegisterUser(body.Name, body.Email, claims.Phone); err != nil {
		if err == auth.ErrUserAlreadyExist {

			c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.JSONMessage("already registered"))
		} else {

			c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.JSONMessage("server error"))
		}
		return
	}
	token, err := h.auth.GenerateJWT(claims.Phone)

	if err != nil {
		h.log.Error().Err(err).Msg("cannot generate jwt")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK,
		helpers.AppendJSONMessage("registration successful", gin.H{"access_token": token}))

}
