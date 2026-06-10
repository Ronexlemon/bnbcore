package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID    *uuid.UUID  `json:"user_id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	TenantID  *uuid.UUID  `json:"tenant_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken   string  `json:"access_token"`
	RefreshToken  string  `json:"refresh_token"`
	RefreshClaims *Claims `json:"claims"`
}

type JwtManager struct {
	secret               []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewJwtManager(secret []byte, accessTokenDuration, refreshTokenDuration time.Duration) *JwtManager {
	return &JwtManager{
		secret:               secret,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

//
func (m *JwtManager) GenerateTokenPair(userID *uuid.UUID, email, role string) (*TokenPair, error) {
	_, access, err := m.generateToken(userID,  email, role, m.accessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshClaims, refresh, err := m.generateToken(userID,"", "", m.refreshTokenDuration)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:   access,
		RefreshToken:  refresh,
		RefreshClaims: refreshClaims,
	}, nil
}


func (m *JwtManager) RefreshAccessToken(refreshToken, email, role string) (*Claims, string, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return nil, "", err
	}
	
	return m.generateToken(claims.UserID, email, role, m.accessTokenDuration)
}

func (m *JwtManager) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwt.Token) (any, error) {
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



func (m *JwtManager) generateToken(
	userID *uuid.UUID,
	email, role string,
	duration time.Duration,
) (*Claims, string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return nil, "", err
	}
	return claims, signed, nil
}