package utils

import (
	"context"
	"go-appointement/storage"
	"os"
	"strconv"
	"time"

	"github.com/kataras/iris/v12/middleware/jwt"
)

var bgContext = context.Background()

func CreateToken(id uint) (*jwt.TokenPair, error) {
	accessTokenSinger := jwt.NewSigner(jwt.HS256, os.Getenv("ACCESS_TOKEN_SECRET"), 24*time.Hour)
	refreshTokenSigner := jwt.NewSigner(jwt.HS256, os.Getenv("REFRSEH_TOKEN_SECRET"), 365*24*time.Hour)

	userID := strconv.FormatUint(uint64(id), 10)

	refreshClaims := jwt.Claims{Subject: userID}

	accessClaims := AccessToken{
		ID: id,
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

type AccessToken struct {
	ID uint `json:"ID"`
}
