package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置CORS头
		c.Header("Access-Control-Allow-Origin", "*") // 在生产环境中建议设置具体域名
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
