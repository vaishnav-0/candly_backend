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
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/ssh"
)

const authPrefix = "auth"
const otpRegenerateTime = 5
const OTPMaxRegeneration = 3
const OTPMaxRetries = 3

var OTPLimitError = errors.New("OTP limit exceeded")
var OTPRetryError = errors.New("OTP regenerate wait time not reached")
var OTPInvalidError = errors.New("OTP invalid")
var OTPTriesExceededError = errors.New("OTP tries reached")

type Auth struct {
	jwt_key interface{}
	jwt_pub interface{}
	rd      *redis.Client
	db      *pgxpool.Pool
}

type JwtClaims struct {
	Role string `json:"role"`
	User string `json:"user"`
	jwt.RegisteredClaims
}

func New(key string, pubKey string, rd *redis.Client, db *pgxpool.Pool) *Auth {
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
		panic(fmt.Errorf("Unable to parse Ed25519 public key: %w", err))
	}

	if err != nil {
		panic(err.Error())
	}

	return &Auth{
		jwt_key: k,
		jwt_pub: ed25519Key,
		rd:      rd,
		db:      db,
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
			return OTPLimitError
		}

		res, _ = a.rd.HGet(ctx, id, "time").Result()

		//if tried validating before generating the key would not exist
		if res != "" {
			t, _ := strconv.ParseInt(res, 10, 64)
			fmt.Println(t - time.Now().Unix())
			if (time.Now().Unix() - t) < otpRegenerateTime {
				return OTPRetryError
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
		return OTPRetryError
	}

	_ = a.rd.HIncrBy(ctx, id, "tries", 1)

	if res[0] == otp {
		a.rd.Del(ctx, id)
		return nil
	} else {
		return OTPInvalidError
	}

}

func (a *Auth) GenerateJWT(phone string) (string, error) {

	token := jwt.New(jwt.SigningMethodEdDSA)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(2 * time.Hour).Unix()
	claims["authorized"] = true
	claims["user"] = "username"
	claims["role"] = "user"

	tokenString, err := token.SignedString(a.jwt_key)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Auth) VerifyJWT(token string) (*JwtClaims, error) {

	tkn, err := jwt.ParseWithClaims(token, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return a.jwt_pub, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := tkn.Claims.(*JwtClaims); ok && tkn.Valid {
		// fmt.Println(claims)
		return claims, nil
	} else {
		return nil, errors.New("Invalid claims")
	}

}
