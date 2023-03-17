package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type APIServer struct {
	Addr   string
	Logger *zap.Logger
}

type HttpError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (s *HttpError) Error() string {
	return s.Message
}

type APIFunc func(context.Context, http.ResponseWriter, *http.Request) error

type RequestInfoKey string

func (s APIServer) MakeHTTPHandler(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := context.WithValue(context.Background(), RequestInfoKey("RequestId"), uuid.NewString())
		if err := f(ctx, w, r); err != nil {
			msgErr := &HttpError{StatusCode: 500}
			if errors.As(err, &msgErr) {
				if err := WriteJSON(w, msgErr.StatusCode, msgErr); err != nil {
					panic(err)
				}
			} else {
				if err := WriteJSON(w, http.StatusInternalServerError, err.Error()); err != nil {
					panic(err)
				}
			}
			s.Logger.Info(err.Error(),
				zap.String("RequestId", ctx.Value(RequestInfoKey("RequestId")).(string)),
				zap.Int("StatusCode", msgErr.StatusCode),
			)
		}
	}
}

func (s APIServer) Run(router *mux.Router) error {
	s.Logger.Info("server is running", zap.String("address", s.Addr))
	return http.ListenAndServe(s.Addr, router)
}

func WriteJSON(w http.ResponseWriter, statusCode int, v any) error {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(v)
}

func Errorf(statusCode int, format string, a ...any) error {
	return &HttpError{
		Message:    fmt.Sprintf(format, a...),
		StatusCode: statusCode,
	}
}
