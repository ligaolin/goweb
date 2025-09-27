package goweb

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtConfig struct {
	Expir  int64  `json:"expir" toml:"expir" yaml:"expir"` // jwt登录过期时间，分钟，1440一天
	Issuer string `json:"issuer" toml:"issuer" yaml:"issuer"`
}

type Claims struct {
	ID   int32  `json:"id"`
	Type string `json:"type"`
	jwt.RegisteredClaims
}

type Jwt struct {
	Config *JwtConfig
}

func NewJwt(cfg *JwtConfig) *Jwt {
	return &Jwt{
		Config: cfg,
	}
}

func (j *Jwt) Set(id int32, types string) (string, error) {
	claims := Claims{
		id,
		types,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.Config.Expir) * time.Minute)),
			Issuer:    j.Config.Issuer,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("AllYourBase"))
}

func (j *Jwt) Get(t string, claims *Claims) error {
	token, err := jwt.ParseWithClaims(t, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("身份信无效或已过期")
		}
		return []byte("AllYourBase"), nil
	})
	if err != nil {
		return err
	}

	if _, ok := token.Claims.(*Claims); ok {
		return nil
	} else {
		return errors.New("解析身份信息失败")
	}
}
