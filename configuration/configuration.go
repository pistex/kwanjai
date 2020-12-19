package configuration

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

var (
	BaseDirectory                string
	Port                         string
	FirebaseProjectID            string
	FrontendURL                  string
	BackendURL                   string
	EmailServicePassword         string
	EmailVerificationLifetime    time.Duration
	JWTAccessTokenSecretKey      string
	JWTRefreshTokenSecretKey     string
	JWTAccessTokenLifetime       time.Duration
	JWTRefreshTokenLifetime      time.Duration
	Context                      context.Context
	DefaultAuthenticationBackend gin.HandlerFunc
	OmisePublicKey               string
	OmiseSecretKey               string
	SQL                          *gorm.DB
	SQLHostname                  string
	SQLUsername                  string
	SQLPassword                  string
	SQLDatabaseName              string
)
