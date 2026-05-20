package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


type Claims struct{
	UserID string `json:"user_id"`
	Email string `json:"email"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct{
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type JwtManager struct{
	secret []byte 
	accessTokenDuration time.Duration
	refreshTokenDuration  time.Duration
}

func NewJwtManager(secret [] byte,accessTokenDuration,refreshTokenDuration time.Duration)*JwtManager{

	return&JwtManager{
		secret: secret,
		accessTokenDuration: accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}

}

func (m *JwtManager) GenerateTokenPair(userID, email, role string) (*TokenPair, error) {
	access, err := m.generateToken(userID, email, role, m.accessTokenDuration)
	if err != nil {
		return nil, err
	}
 
	refresh, err := m.generateToken(userID, "", "", m.refreshTokenDuration)
	if err != nil {
		return nil, err
	}
 
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (m *JwtManager) generateToken(userId,email,role string,duration time.Duration)(string,error){

	now:=time.Now()

	claims:=&Claims{
		UserID: userId,
		Email: email,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userId,
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token:=jwt.NewWithClaims(jwt.SigningMethodES256,claims)
	return token.SignedString(m.secret)

}

func (m *JwtManager) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return m.secret, nil
		},
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
 
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func (m *JwtManager) RefreshAccessToken(refreshToken, email, role string) (string, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}
	return m.generateToken(claims.UserID, email, role, m.accessTokenDuration)
}
 