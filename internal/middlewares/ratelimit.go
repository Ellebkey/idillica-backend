// ratelimit.go: token bucket por IP con x/time/rate — un mapa protegido por
// mutex más una goroutine de limpieza.
package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ipLimiter struct {
	mu      sync.Mutex
	entries map[string]*limiterEntry
	rate    rate.Limit
	burst   int
}

func newIPLimiter(r rate.Limit, burst int) *ipLimiter {
	l := &ipLimiter{
		entries: make(map[string]*limiterEntry),
		rate:    r,
		burst:   burst,
	}
	// Evict idle IPs so the map never grows unbounded.
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			l.mu.Lock()
			for ip, entry := range l.entries {
				if time.Since(entry.lastSeen) > 30*time.Minute {
					delete(l.entries, ip)
				}
			}
			l.mu.Unlock()
		}
	}()
	return l
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.entries[ip]
	if !ok {
		entry = &limiterEntry{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.entries[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

// rateLimiter builds a middleware allowing `requests` per `window` per IP
// (token bucket: burst = requests, sustained rate = requests/window).
// Deshabilitado en el entorno de test.
func rateLimiter(env string, requests int, window time.Duration) gin.HandlerFunc {
	if env == "test" {
		return func(c *gin.Context) { c.Next() }
	}

	limiter := newIPLimiter(rate.Every(window/time.Duration(requests)), requests)

	return func(c *gin.Context) {
		if !limiter.allow(c.ClientIP()) {
			// Mismo cuerpo JSON que el resto de errores del API
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "TOO_MANY_REQUESTS",
					"message": "Too many requests, please try again later",
					"status":  429,
				},
			})
			return
		}
		c.Next()
	}
}

// AuthRateLimiter: 10 requests / 15 min para los endpoints de auth.
func AuthRateLimiter(env string) gin.HandlerFunc {
	return rateLimiter(env, 10, 15*time.Minute)
}

// APIRateLimiter: 100 requests / minuto para todo el API.
func APIRateLimiter(env string) gin.HandlerFunc {
	return rateLimiter(env, 100, time.Minute)
}
