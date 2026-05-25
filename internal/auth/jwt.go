package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/satori/go.uuid"
)


type Claims struct{
	UserID uuid.UUID `json:"user_id"`
	Email string `json:"email"`
	Role string `json:"role"`
	TenantID string `json:"tenant_id"`  
    Subdomain string `json:"subdomain"` 
	jwt.RegisteredClaims
}

type TokenPair struct{
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	RefreshClaims *Claims `json:"claims"`
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

func (m *JwtManager) GenerateTokenPair(userID uuid.UUID, email, role string,subdomain string) (*TokenPair, error) {
	_,access, err := m.generateToken(userID, email, role, m.accessTokenDuration ,subdomain)
	if err != nil {
		return nil, err
	}
 
	claims,refresh, err := m.generateToken(userID, "", "", m.refreshTokenDuration,subdomain)
	if err != nil {
		return nil, err
	}
 
	return &TokenPair{AccessToken: access, RefreshToken: refresh,RefreshClaims:claims }, nil
}

func (m *JwtManager) generateToken(userId uuid.UUID,email,role string,duration time.Duration,subdomain string)(*Claims, string, error){

	now:=time.Now()

	claims:=&Claims{
		UserID: userId,
		Email: email,
		Role: role,
		Subdomain: subdomain,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userId.String(),
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) 
	signedStr, err := token.SignedString(m.secret)
	if err != nil {
		return nil, "", err
	}
return claims,signedStr,err

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

func (m *JwtManager) RefreshAccessToken(refreshToken, email, role string,subdomain string) (*Claims,string, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return nil,"", err
	}
	return m.generateToken(claims.UserID, email, role, m.accessTokenDuration,subdomain)
}
 