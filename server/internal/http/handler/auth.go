package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

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

func GenerateOTP(a *auth.Auth, log *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		body := GenerateOTPBody{}
		if err := c.BindJSON(&body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, helpers.JSONMessage("mobile number is required"))
			return
		}

		otp, err := a.GenerateOTP()

		if err != nil {
			log.Error().Err(err).Msg("cannot generate OPT")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		err = a.StoreOTP(body.Phone, otp)

		if err == auth.ErrOTPLimit {
			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("OTP limit exceeded"))
			return
		} else if err == auth.ErrOTPRetry {
			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("OTP wait time not reached"))
			return
		}

		c.JSON(http.StatusOK, gin.H{"otp": otp})

	}

}

func VerifyOTP(a *auth.Auth, log *zerolog.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {

		body := VerifyOTPBody{}
		if err := c.BindJSON(&body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, helpers.JSONMessage("mobile number and otp is required"))
			return
		}

		if err := a.VerifyOTP(body.Phone, body.Otp); err == nil {

			access, refresh, err := a.GenerateTokens(body.Phone)

			if err != nil {
				log.Error().Err(err).Msg("cannot generate access token")
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			if refresh == "" {

				c.JSON(http.StatusOK, gin.H{"access_token": access})

			} else {

				c.JSON(http.StatusOK, gin.H{"access_token": access, "refresh_token": refresh})

			}

		} else if err == auth.ErrOTPInvalid {

			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("invalid otp"))

		} else if err == auth.ErrOTPRetry {

			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("otp retries exceeded"))

		}

	}

}

type RegisterUserBody struct {
	Name  string `json:"name"  binding:"required"`
	Email string `json:"email" binding:"email"`
}

func RegisterUser(a *auth.Auth, log *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		if err := a.RegisterUser(body.Name, body.Email, claims.Phone); err != nil {
			if err == auth.ErrUserAlreadyExist {

				c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.JSONMessage("already registered"))
			} else {

				c.AbortWithStatusJSON(http.StatusInternalServerError, helpers.JSONMessage("server error"))
			}
			return
		}
		token, refresh, err := a.GenerateTokens(claims.Phone)

		if err != nil {
			log.Error().Err(err).Msg("cannot generate jwt")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK,
			helpers.AppendJSONMessage("registration successful", gin.H{"access_token": token, "refresh_token": refresh}))

	}

}

type RefreshTokenBody struct {
	Token string `json:"token"  binding:"required"`
}

func RefreshToken(a *auth.Auth, log *zerolog.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {

		body := RefreshTokenBody{}
		if err := c.BindJSON(&body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, helpers.GenerateValidationError(err))
			return
		}

		access, err := a.AccessFromRefresh(body.Token)

		if err != nil {
			log.Error().Err(err).Msg("cannot generate access token")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, gin.H{"access_token": access})

	}
}
