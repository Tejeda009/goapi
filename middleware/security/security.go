package securiy

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	conf "github.com/Tejeda009/goapi/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

func Auth(t *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		//fmt.Printf("token: %v", token) crashed everything

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header not found"})
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")
		dec, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(conf.Key), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Alg()}))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not a valid JWT"})
			return
		}
		switch {
		case dec.Valid:
			if claims, ok := dec.Claims.(jwt.MapClaims); ok {

				if claims["iss"] == conf.Iss {
					if user, ok := claims["user"].(string); ok {
						*t = user
						c.Next()
						return
					}

					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "claim user not found"})
					return
				}

				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid issuer"})
				return
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "wrong claims my friend"})
			return
		case errors.Is(err, jwt.ErrTokenMalformed):
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Malformed Token"})
			return
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Toker Signature Invalid"})
			return
		case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token Expired"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
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

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Authorization, Content-Length")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func RateLimiter() gin.HandlerFunc {
	type client struct {
		limiter *rate.Limiter
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
		}

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
