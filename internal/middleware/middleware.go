package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// responseWriter is a wrapper to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging middleware logs HTTP requests
func Logging(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			
			next.ServeHTTP(wrapped, r)
			
			logger.WithFields(logrus.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     wrapped.statusCode,
				"duration":   time.Since(start),
				"ip":         r.RemoteAddr,
				"user_agent": r.UserAgent(),
			}).Info("HTTP request")
		})
	}
}

// Recovery middleware recovers from panics
func Recovery(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.WithFields(logrus.Fields{
						"error": err,
						"path":  r.URL.Path,
					}).Error("Panic recovered")
					
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// GRPCLogging logs gRPC requests
func GRPCLogging(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		
		resp, err := handler(ctx, req)
		
		logger.WithFields(logrus.Fields{
			"method":   info.FullMethod,
			"duration": time.Since(start),
			"error":    err,
		}).Info("gRPC request")
		
		return resp, err
	}
}