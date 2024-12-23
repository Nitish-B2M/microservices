package utils

import "net/http"

type CustomResponseWriter struct {
	http.ResponseWriter
	headerWritten bool
	statusCode    int
}

func (cw *CustomResponseWriter) WriteHeader(statusCode int) {
	// If the header has already been written, do not rewrite it
	if cw.headerWritten {
		return
	}
	// Ensure that the status code is set to the custom code passed
	if statusCode == 0 {
		statusCode = http.StatusOK // Default to StatusOK if no status code is passed
	}
	cw.statusCode = statusCode
	cw.ResponseWriter.WriteHeader(statusCode)
	cw.headerWritten = true
}

func (cw *CustomResponseWriter) Write(p []byte) (n int, err error) {
	// Ensure the header is written before sending the body
	if !cw.headerWritten {
		cw.WriteHeader(http.StatusOK) // Default to StatusOK if no header set
	}
	return cw.ResponseWriter.Write(p)
}

func NewCustomResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{ResponseWriter: w}
}
