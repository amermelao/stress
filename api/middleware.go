package main

import (
	"log/slog"
	"net/http"
	"time"
)

type status struct {
	http.ResponseWriter
	statusCode int
}

func (s *status) WriteHeader(statusCode int) {
	s.ResponseWriter.WriteHeader(statusCode)
	s.statusCode = statusCode
}

func logRequests(next http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		tNow := time.Now()
		statusCode := &status{statusCode: http.StatusOK, ResponseWriter: writer}
		next.ServeHTTP(statusCode, request)

		slog.Info("incoming",
			"status", statusCode.statusCode,
			"url", request.URL,
			"method", request.Method,
			"time", time.Since(tNow).String(),
		)
	})
}

func secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") == "Bearer 5Mxrg3TkCRq4aMy4PyO8QYA7BiWUqHy9fPlVbSruAlDpGj10ry4mgbbetL79M12S" {
			next.ServeHTTP(writer, request)
		} else {
			writer.WriteHeader(http.StatusUnauthorized)
		}
	})
}
