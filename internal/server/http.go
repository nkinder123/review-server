package server

import (
	"github.com/go-kratos/kratos/v2/middleware/validate"
	v1 "review-server/api/review/v1"
	"review-server/internal/conf"
	"review-server/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type Response struct {
	Code    int
	Message string
	Data    interface{}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, review *service.ReviewService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			validate.Validator(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	opts = append(opts, http.ResponseEncoder(SuccessResponse))
	opts = append(opts, http.ErrorEncoder(ErrorResponseEncoder))
	srv := http.NewServer(opts...)
	v1.RegisterReviewHTTPServer(srv, review)
	return srv
}
