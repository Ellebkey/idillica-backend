// requestid.go: correlación por request. Toma X-Request-Id del cliente (si
// es válido) o genera uno, lo guarda en el context de la petición y lo
// devuelve en la respuesta. Debe registrarse primero para que todo lo que
// venga después (logs incluidos) lleve el id.
package middlewares

import (
	"github.com/gin-gonic/gin"

	"idilica-backend-go/internal/reqid"
)

// RequestID inyecta el id de correlación en el context y lo refleja en la
// respuesta vía el header X-Request-Id.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := reqid.Sanitize(c.GetHeader(reqid.HeaderName))
		c.Request = c.Request.WithContext(reqid.WithID(c.Request.Context(), id))
		c.Header(reqid.HeaderName, id)
		c.Next()
	}
}
