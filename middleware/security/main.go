package securiy

import (
	"time"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

var key string = "test" // TODO add env
var iss string = "test" // TODO add env
var TTL int = 0 // TODO add env

func Auth( t *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Bearer")

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Bearer header not found"})
		}

		dec, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return key, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not a valid JWT"})
		}

		if claims, ok := dec.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["expireAt"].(float64); ok {
				expireTime := time.Unix(int64(exp), 0)

				if claims["iss"] == iss && expireTime.Sub(time.Now()) > TTL {
					if user, ok := claims["user"].(string); ok {
						*t = user
						c.Next()
					}

					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error":"claim user not found"})
				}

				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "claim iss or expire time invalid"})
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "claim expireAt not found"})
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "wrong claims my friend"})
		}
	}

}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		c.Next()
	}
}

func RateLimiter() gin.HandlerFunc {
	type client struct {
		limiter *rate.limiter
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()

		if _, exists := clients[ip]; !exists {
			clients[ip] = &client{limiter: rate.NewLimiter(10, 20)}

			cl := clients[ip]
			mu.Unlock()

			if !cl.limiter.Allow() {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "rate limit exceeded",
				})
				return
			}
			c.Next()
		}
	}
}
