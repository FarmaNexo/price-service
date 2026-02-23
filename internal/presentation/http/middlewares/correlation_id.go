// internal/presentation/http/middlewares/correlation_id.go
package middlewares

import (
	"context"
	"net/http"

	"github.com/farmanexo/price-service/pkg/mediator"
	"github.com/google/uuid"
)

// CorrelationID middleware que agrega correlation ID a cada request
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		w.Header().Set("X-Correlation-ID", correlationID)

		ctx := mediator.WithValue(r.Context(), mediator.CorrelationKey, correlationID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetCorrelationID obtiene el correlation ID del contexto
func GetCorrelationID(ctx context.Context) string {
	if val := ctx.Value(mediator.CorrelationKey); val != nil {
		if corrID, ok := val.(string); ok {
			return corrID
		}
	}
	return ""
}
