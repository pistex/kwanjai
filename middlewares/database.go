package middlewares

import (
	"github.com/gin-gonic/gin"
	"kwanjai/configuration"
)

func MySQL() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		ginContext.Set("db", configuration.SQL)
	}
}

func Redis() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		ginContext.Set("cache", configuration.Redis)
	}
}
