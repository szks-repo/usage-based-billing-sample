package httplib

import "net/http"

type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (ww *ResponseWriterWrapper) StatusCode() int {
	return ww.statusCode
}

func (ww *ResponseWriterWrapper) BytesWritten() int {
	return ww.bytesWritten
}

func (w *ResponseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseWriterWrapper) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.bytesWritten += n
	return n, err
}
