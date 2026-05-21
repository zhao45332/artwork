package server

import (
	artworkv1 "artwork/api/artwork/v1"
	authv1 "artwork/api/auth/v1"
	userv1 "artwork/api/user/v1"
	"artwork/internal/conf"
	"artwork/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, artwork *service.ArtworkService, auth *service.AuthService, user *service.UserService, authMiddleware middleware.Middleware, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			authMiddleware,
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	authv1.RegisterAuthServiceServer(srv, auth)
	artworkv1.RegisterArtworkServiceServer(srv, artwork)
	userv1.RegisterUserServiceServer(srv, user)
	return srv
}
