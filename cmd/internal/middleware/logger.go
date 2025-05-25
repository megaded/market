package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/megaded/market/cmd/internal/logger"
	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type responseLogWriter struct {
	http.ResponseWriter
	data responseData
}

func (r *responseLogWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.data.size += size
	return size, err
}

func (r *responseLogWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.data.status = statusCode
}

func (r *responseLogWriter) logResponse() {
	logger.Log.Info("Response status", zap.Int("status", r.data.status))
	logger.Log.Info("Response size", zap.Int("size", r.data.size))
	logger.Log.Info("Response hash", zap.String("hash", r.Header().Get("HashSHA256")))
}

func Logger(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := responseLogWriter{
			ResponseWriter: w,
		}
		var requestBody []byte
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Info(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			requestBody = bodyBytes
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		logger.Log.Info("Request uri", zap.String("uri", r.RequestURI))
		logger.Log.Info("Request method", zap.String("method", r.Method))
		logger.Log.Info("Request duration", zap.Duration("duration", duration))
		logger.Log.Info("Request body", zap.String("body", string(requestBody)))
		logger.Log.Info("Request Content-type", zap.String("content-type", r.Header.Get("Content-Type")))
		logger.Log.Info("Request Content-Encoding", zap.String("Content-Encoding", r.Header.Get("Content-Encoding")))
		logger.Log.Info("Request Hash", zap.String("Hash", r.Header.Get("HashSHA256")))

		lw.logResponse()

	}

	return http.HandlerFunc(logFn)
}
