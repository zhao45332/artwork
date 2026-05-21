package service

import (
	"context"
	"strings"

	authv1 "artwork/api/auth/v1"
	"artwork/internal/biz"
	"artwork/internal/pkg/xctx"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthService struct {
	authv1.UnimplementedAuthServiceServer

	uc  *biz.AuthUsecase
	log *log.Helper
}

func NewAuthService(uc *biz.AuthUsecase, logger log.Logger) *AuthService {
	return &AuthService{uc: uc, log: log.NewHelper(logger)}
}

func (s *AuthService) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthReply, error) {
	user, accessToken, refreshToken, expiresIn, err := s.uc.Register(ctx, req, clientIPFromContext(ctx))
	if err != nil {
		return nil, err
	}
	return s.toAuthReply(user, accessToken, refreshToken, expiresIn), nil
}

func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthReply, error) {
	user, accessToken, refreshToken, expiresIn, err := s.uc.Login(ctx, req.Account, req.Password, clientIPFromContext(ctx))
	if err != nil {
		return nil, err
	}
	return s.toAuthReply(user, accessToken, refreshToken, expiresIn), nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.AuthReply, error) {
	user, accessToken, refreshToken, expiresIn, err := s.uc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return s.toAuthReply(user, accessToken, refreshToken, expiresIn), nil
}

func (s *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutReply, error) {
	sessionID, _ := xctx.SessionIDFromContext(ctx)
	if err := s.uc.Logout(ctx, req.RefreshToken, sessionID); err != nil {
		return nil, err
	}
	return &authv1.LogoutReply{Success: true}, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, _ *authv1.GetCurrentUserRequest) (*authv1.GetCurrentUserReply, error) {
	userID, ok := xctx.UserIDFromContext(ctx)
	if !ok {
		return nil, biz.ErrUnauthenticated
	}
	user, err := s.uc.GetCurrentUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &authv1.GetCurrentUserReply{User: s.toUserInfo(user)}, nil
}

func (s *AuthService) toAuthReply(user *biz.User, accessToken, refreshToken string, expiresIn int64) *authv1.AuthReply {
	return &authv1.AuthReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User:         s.toUserInfo(user),
	}
}

func (s *AuthService) toUserInfo(user *biz.User) *authv1.UserInfo {
	info := &authv1.UserInfo{
		Id:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		AvatarUrl: user.AvatarURL,
		Bio:       user.Bio,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
	}
	if !user.CreatedAt.IsZero() {
		info.CreatedAt = timestamppb.New(user.CreatedAt)
	}
	if !user.UpdatedAt.IsZero() {
		info.UpdatedAt = timestamppb.New(user.UpdatedAt)
	}
	return info
}

func clientIPFromContext(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok || tr.Kind() != transport.KindHTTP {
		return ""
	}
	forwarded := strings.TrimSpace(tr.RequestHeader().Get("X-Forwarded-For"))
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	return strings.TrimSpace(tr.RequestHeader().Get("X-Real-IP"))
}
