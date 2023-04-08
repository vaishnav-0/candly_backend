package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"candly/internal/auth"
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
		c.AbortWithStatusJSON(http.StatusBadRequest, JSONMessage("mobile number is required"))
		return
	}

	otp, err := h.auth.GenerateOTP()

	if err != nil {
		h.log.Error().Err(err).Msg("cannot generate OPT")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = h.auth.StoreOTP(body.Phone, otp)

	if err == auth.OTPLimitError {
		c.AbortWithStatusJSON(http.StatusUnauthorized, JSONMessage("OTP limit exceeded"))
		return
	} else if err == auth.OTPRetryError {
		c.AbortWithStatusJSON(http.StatusUnauthorized, JSONMessage("OTP wait time not reached"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"otp": otp})

}

func (h *Handlers) VerifyOTP(c *gin.Context) {

	body := VerifyOTPBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, JSONMessage("mobile number and otp is required"))
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

	} else if err == auth.OTPInvalidError {

		c.AbortWithStatusJSON(http.StatusUnauthorized, JSONMessage("invalid otp"))

	} else if err == auth.OTPRetryError {

		c.AbortWithStatusJSON(http.StatusUnauthorized, JSONMessage("otp retries exceeded"))

	}

}
