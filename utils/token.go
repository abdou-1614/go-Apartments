package utils

import (
	"context"
	"go-appointement/model"
	"go-appointement/storage"
	"os"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
)

var bgContext = context.Background()

func CreateForgetPasswordToken(id uint, email string) (string, error) {
	signer := jwt.NewSigner(jwt.HS256, os.Getenv("EMAIL_SECRET_TOKEN"), 10*time.Minute)

	claims := ForgetPasswordToken{
		ID:    id,
		Email: email,
	}

	token, err := signer.Sign(claims)

	if err != nil {
		return "", err
	}

	return string(token), nil
}

func CreateToken(id uint, role model.UserRole) (*jwt.TokenPair, error) {
	accessTokenSinger := jwt.NewSigner(jwt.HS256, os.Getenv("ACCESS_TOKEN_SECRET"), 24*time.Hour)
	refreshTokenSigner := jwt.NewSigner(jwt.HS256, os.Getenv("REFRSEH_TOKEN_SECRET"), 365*24*time.Hour)

	userID := strconv.FormatUint(uint64(id), 10)

	refreshClaims := jwt.Claims{Subject: userID}

	accessClaims := AccessToken{
		ID:   id,
		ROLE: role,
	}

	accessToken, err := accessTokenSinger.Sign(accessClaims)

	if err != nil {
		return nil, err
	}

	refreshToken, err := refreshTokenSigner.Sign(refreshClaims)

	if err != nil {
		return nil, err
	}

	var tokenPair jwt.TokenPair

	tokenPair.AccessToken = accessToken
	tokenPair.RefreshToken = refreshToken

	storage.Redis.Set(bgContext, string(refreshToken), "true", 365*24*time.Hour+5*time.Minute)

	return &tokenPair, nil

}

func RefreshToken(ctx iris.Context) {
	token := jwt.GetVerifiedToken(ctx)
	tokenStr := string(token.Token)

	var claims AccessToken
	claimsErr := token.Claims(&claims)
	if claimsErr != nil {
		CreateInternalServerError(ctx)
		return
	}

	// Extract the role from the claims
	originalRole := claims.ROLE

	validateToken, tokenErr := storage.Redis.Get(bgContext, tokenStr).Result()

	if tokenErr != nil {
		CreateError(iris.StatusNotFound, "Token Not Found", "Token Not Found", ctx)
		return
	}

	if validateToken != "true" {
		CreateInternalServerError(ctx)
		return
	}

	storage.Redis.Del(bgContext, tokenStr)

	userID, parseID := strconv.ParseUint(token.StandardClaims.Subject, 10, 32)

	if parseID != nil {
		CreateInternalServerError(ctx)
		return
	}

	tokenPair, tokenPairErr := CreateToken(uint(userID), originalRole)

	if tokenPairErr != nil {
		CreateInternalServerError(ctx)
		return
	}

	ctx.JSON(iris.Map{
		"accessToken":  string(tokenPair.AccessToken),
		"refreshToken": string(tokenPair.RefreshToken),
	})
}

type AccessToken struct {
	ID   uint           `json:"ID"`
	ROLE model.UserRole `json:"ROLE"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type ForgetPasswordToken struct {
	ID    uint   `json:"ID"`
	Email string `json:"email"`
}
