package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"strconv"

	// "strings"

	// "errors"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh"

	"candly/internal/db/queries"
)

const authPrefix = "auth"
const authRefreshPrefix = "ref:"
const otpRegenerateTime = 5
const OTPMaxRegeneration = 3
const OTPMaxRetries = 3

var ErrOTPLimit = errors.New("OTP limit exceeded")
var ErrOTPRetry = errors.New("OTP regenerate wait time not reached")
var ErrOTPInvalid = errors.New("OTP invalid")
var ErrOTPTriesExceeded = errors.New("OTP tries reached")

var ErrUserAlreadyExist = errors.New("user already exist")
var ErrUserUnregistered = errors.New("user already exist")

var ErrInvalidRefreshToken = errors.New("invalid refresh token")

type Auth struct {
	jwt_key interface{}
	jwt_pub interface{}
	rd      *redis.Client
	db      *pgxpool.Pool
	log     *zerolog.Logger
}

type JwtUserClaims struct {
	Roles []string `json:"roles"`
	User  string   `json:"user"`
	jwt.RegisteredClaims
}

type JwtNewUserClaims struct {
	Roles []string `json:"roles"`
	Phone string   `json:"phone"`
	jwt.RegisteredClaims
}

func New(key string, pubKey string, rd *redis.Client, db *pgxpool.Pool, log *zerolog.Logger) *Auth {
	keyData, err := os.ReadFile(key)

	if err != nil {
		panic(err.Error())
	}

	k, err := ssh.ParseRawPrivateKey(keyData)
	if err != nil {
		panic(err.Error())
	}

	pKey, err := os.ReadFile(pubKey)
	if err != nil {
		panic(err.Error())
	}

	ed25519Key, err := jwt.ParseEdPublicKeyFromPEM(pKey)
	if err != nil {
		panic(fmt.Errorf("unable to parse Ed25519 public key: %w", err))
	}

	if err != nil {
		panic(err.Error())
	}

	return &Auth{
		jwt_key: k,
		jwt_pub: ed25519Key,
		rd:      rd,
		db:      db,
		log:     log,
	}
}

func (*Auth) GenerateOTP() (string, error) {
	seed := "012345679"
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(b); i++ {
		b[i] = seed[int(b[i])%len(seed)]
	}

	return string(b), nil
}

func (a *Auth) StoreOTP(phone string, otp string) error {
	ctx := context.Background()
	id := authPrefix + phone

	exist, _ := a.rd.Exists(ctx, id).Result()

	if exist == 1 {
		res, _ := a.rd.HGet(ctx, id, "generations").Result()

		tries, _ := strconv.ParseInt(res, 10, 64)

		if tries == OTPMaxRegeneration {
			return ErrOTPLimit
		}

		res, _ = a.rd.HGet(ctx, id, "time").Result()

		//if tried validating before generating the key would not exist
		if res != "" {
			t, _ := strconv.ParseInt(res, 10, 64)
			fmt.Println(t - time.Now().Unix())
			if (time.Now().Unix() - t) < otpRegenerateTime {
				return ErrOTPRetry
			}
		}

	}

	_, err := a.rd.HSet(ctx, id, "otp", otp, "time", time.Now().Unix()).Result()
	if err != nil {
		return fmt.Errorf("failed to store otp. %w", err)
	}
	a.rd.Expire(ctx, id, time.Second*90).Result()

	_ = a.rd.HIncrBy(ctx, id, "generations", 1)

	return nil
}

func (a *Auth) VerifyOTP(phone string, otp string) error {
	ctx := context.Background()
	id := authPrefix + phone
	res, _ := a.rd.HMGet(ctx, id, "otp", "tries").Result()

	tr, _ := res[1].(string)
	tries, _ := strconv.ParseInt(tr, 10, 64)

	if tries == OTPMaxRetries {
		return ErrOTPRetry
	}

	_ = a.rd.HIncrBy(ctx, id, "tries", 1)

	if res[0] == otp {
		a.rd.Del(ctx, id)
		return nil
	} else {
		return ErrOTPInvalid
	}

}

func (a *Auth) GenerateRefreshToken(id string) (string, error) {

	token := jwt.New(jwt.SigningMethodEdDSA)
	claims := token.Claims.(jwt.MapClaims)

	guid := xid.New().String()

	exp := time.Now().Add(12 * time.Hour)

	claims["exp"] = exp.Unix()
	claims["sub"] = id
	claims["jti"] = guid

	ctx := context.Background()
	_, err := a.rd.Set(ctx, authRefreshPrefix+guid, true, time.Until(exp)).Result()

	if err != nil {
		return "", err
	}

	tokenString, err := token.SignedString(a.jwt_key)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Auth) GenerateTokens(phone string) (string, string, error) {

	q := queries.New(a.db)
	ctx := context.Background()

	user, err := q.GetUserFromPhone(ctx, phone)

	if err == pgx.ErrNoRows {
		token := jwt.New(jwt.SigningMethodEdDSA)
		claims := token.Claims.(jwt.MapClaims)

		claims["exp"] = time.Now().Add(20 * time.Minute).Unix()
		claims["roles"] = []string{"new"}
		claims["phone"] = phone

		acc, err := token.SignedString(a.jwt_key)
		return acc, "", err

	} else if err == nil {

		acc, err := a.GenerateUserJWT(user)

		if err != nil {
			a.log.Err(err).Msg("error generating access token")
		}

		id := strconv.Itoa(int(user.ID))
		ref, err := a.GenerateRefreshToken(string(id))

		if err != nil {

			a.log.Err(err).Msg("error generating access token")
		}

		return acc, ref, err

	}

	return "", "", err

}

func (a *Auth) GenerateUserJWT(user queries.User) (string, error) {

	token := jwt.New(jwt.SigningMethodEdDSA)
	claims := token.Claims.(jwt.MapClaims)

	claims["exp"] = time.Now().Add(2 * time.Hour).Unix()
	claims["user"] = user.Name
	claims["sub"] = user.ID
	claims["roles"] = []string{"user"}

	tokenString, err := token.SignedString(a.jwt_key)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Auth) VerifyRefreshToken(token string) (*jwt.RegisteredClaims, error) {

	tkn, err := a.VerifyJWT(token, &jwt.RegisteredClaims{})

	if err != nil {
		return nil, err
	}

	if claims, ok := tkn.Claims.(*jwt.RegisteredClaims); ok && tkn.Valid {

		res, err := a.rd.Exists(context.Background(), authRefreshPrefix+claims.ID).Result()

		if err != nil {
			return nil, err
		}

		if res != 1 {
			return nil, ErrInvalidRefreshToken
		}

		return claims, nil

	} else {

		return nil, errors.New("invalid claims")

	}

}

func (a *Auth) RevokeRefresh(token string) error {
	tkn, err := a.VerifyJWT(token, &jwt.RegisteredClaims{})

	if err != nil {
		return err
	}

	if claims, ok := tkn.Claims.(*jwt.RegisteredClaims); ok && tkn.Valid {
		_, err := a.rd.Del(context.Background(), authRefreshPrefix+claims.ID).Result()

		if err != nil {
			return err
		}

	}

	return nil
}

func (a *Auth) AccessFromRefresh(token string) (string, error) {

	cl, err := a.VerifyRefreshToken(token)

	if err != nil {
		return "", err
	}

	q := queries.New(a.db)
	ctx := context.Background()
	sub, err := strconv.ParseInt(cl.Subject, 10, 64)

	if err != nil {

		return "", err
	}
	user, err := q.GetUser(ctx, sub)

	if err != nil {
		return "", err
	}

	return a.GenerateUserJWT(user)

}

func (a *Auth) VerifyJWT(token string, claims jwt.Claims) (*jwt.Token, error) {

	return jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return a.jwt_pub, nil
	})

}

func (a *Auth) VerifyUserJWT(token string) (*JwtUserClaims, error) {

	tkn, err := a.VerifyJWT(token, &JwtUserClaims{})

	if err != nil {
		return nil, err
	}

	if claims, ok := tkn.Claims.(*JwtUserClaims); ok && tkn.Valid {

		return claims, nil

	} else {

		return nil, errors.New("invalid claims")

	}

}

func (a *Auth) VerifyNewUserJWT(token string) (*JwtNewUserClaims, error) {

	tkn, err := a.VerifyJWT(token, &JwtNewUserClaims{})

	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	if claims, ok := tkn.Claims.(*JwtNewUserClaims); ok && tkn.Valid {

		return claims, nil

	} else {

		return nil, errors.New("invalid claims")

	}

}

func (a *Auth) RegisterUser(name string, email string, phone string) error {

	q := queries.New(a.db)
	ctx := context.Background()
	nameText := pgtype.Text{}
	nameText.Scan(name)

	emailText := pgtype.Text{}
	emailText.Scan(name)

	err := q.InsertUser(ctx, queries.InsertUserParams{
		Name:  nameText,
		Phone: phone,
		Email: emailText,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrUserAlreadyExist
			}
		}
	}

	return err

}
