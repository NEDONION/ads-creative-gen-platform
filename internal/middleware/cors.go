package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		// 允许任意来源：在带凭证场景下不能返回 "*"，所以回显请求的 Origin
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Max-Age", "7200") // 预检请求缓存时间
			c.AbortWithStatus(204)                     // No Content
			return
		}

		c.Next()
	}
}
