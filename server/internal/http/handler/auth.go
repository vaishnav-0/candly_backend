package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"candly/internal/auth"
	"candly/internal/http/helpers"
)

type GenerateOTPBody struct {
	Phone string `json:"phone" binding:"required"`
}

type GenerateOTPResp struct{
	Otp string `json:"otp"`
}

// GenerateOTP
//
//	@Summary		Generate OTP
//	@Description	Generate an OTP for authentication
//	@Tags			auth
//	@ID				genOTP
//	@Produce		json	
//	@Param			body	body		GenerateOTPBody	true	"Phone number"
//	@Success		200		{object}	GenerateOTPResp
//	@Failure		400		{object}	helpers.ValidationError
//	@Failure		401		{object}	helpers.HTTPMessage
//	@Failure		500
//	@Router			/auth/generateOTP [post]
func GenerateOTP(a *auth.Auth, log *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		body := GenerateOTPBody{}
		if err := c.BindJSON(&body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, helpers.GenerateValidationError(err))
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

		c.JSON(http.StatusOK, GenerateOTPResp{otp})

	}

}



type VerifyOTPBody struct {
	Phone string `json:"phone"`
	Otp   string `json:"otp"`
}


type VerifyOTPRes struct {
	Refresh_token string `json:"refresh_token"`
	Access_token string `json:"access_token"`
}

// ValidateOTP
//
//	@Summary		Validate OTP
//	@Description	Validate an OTP and generate tokens
//	@Tags			auth
//	@ID				valOTP
//	@Produce		json	
//	@Param			body	body		VerifyOTPBody	true	"phone and otp"
//	@Success		200		{object}	VerifyOTPRes
//	@Failure		400		{object}	helpers.ValidationError
//	@Failure		401		{object}	helpers.HTTPMessage
//	@Failure		500
//	@Router			/auth/validate [post]
func VerifyOTP(a *auth.Auth, log *zerolog.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {

		body := VerifyOTPBody{}
		if err := c.BindJSON(&body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, helpers.GenerateValidationError(err))
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

// RegisterUser
//
//	@Summary		Register user
//	@Description	Register a new user
//	@Tags			auth
//	@ID				regUser
//	@Produce		json	
//	@Param			body	body		RegisterUserBody	true	"User details"
//	@Success		200		{object}	VerifyOTPRes
//	@Failure		400		{object}	helpers.ValidationError
//	@Failure		401		{object}	helpers.HTTPMessage
//	@Failure		500
//	@Router			/auth/register [post]
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

type RefreshTokenRes struct {
	Refresh_token string `json:"refresh_token"`
	Access_token string `json:"access_token"`
}
type RefreshTokenBody struct {
	Token string `json:"token"  binding:"required"`
}

// RefreshToken
//
//	@Summary		Refresh token
//	@Description	Refresh access token
//	@Tags			auth
//	@ID				refTkn
//	@Produce		json	
//	@Param			body	body		RefreshTokenBody	true	"refresh token"
//	@Success		200		{object}	RefreshTokenRes
//	@Failure		400		{object}	helpers.ValidationError
//	@Failure		401		{object}	helpers.HTTPMessage
//	@Failure		500
//	@Router			/auth/refresh [post]
func RefreshToken(a *auth.Auth, log *zerolog.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {

		body := RefreshTokenBody{}
		if err := c.BindJSON(&body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, helpers.GenerateValidationError(err))
			return
		}

		access, err := a.AccessFromRefresh(body.Token)

		if err == auth.ErrInvalidRefreshToken{
			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage(err.Error()))
			return
		}

		if err != nil {
			log.Error().Err(err).Msg("cannot generate access token")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, gin.H{"access_token": access})

	}
}

// RevokeRefreshToken
//
//	@Summary		Revoke refresh token
//	@Description	Revoke the given refresh token
//	@Tags			auth
//	@ID				revRefTkn
//	@Produce		json	
//	@Param			body	body		RefreshTokenBody	true	"refresh token"
//	@Success		200		{object}	helpers.HTTPMessage
//	@Failure		400		{object}	helpers.ValidationError
//	@Failure		401		{object}	helpers.HTTPMessage
//	@Failure		500
//	@Router			/auth/revoke [post]
func RevokeRefreshToken(a *auth.Auth, log *zerolog.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {

		body := RefreshTokenBody{}
		if err := c.BindJSON(&body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, helpers.GenerateValidationError(err))
			return
		}

		err := a.RevokeRefresh(body.Token)

		if err != nil {
			log.Error().Err(err).Msg("cannot revoke refresh token")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, helpers.JSONMessage("successful"))

	}
}