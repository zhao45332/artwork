package biz

import (
	"context"
	"strings"
	"time"

	apiauth "artwork/api/auth/v1"
	pkgAuth "artwork/internal/pkg/auth"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

const (
	UserRoleNormal = 1
	UserRoleAdmin  = 9

	UserStatusActive = 1
	UserStatusMuted  = 2
	UserStatusBanned = 3
)

var (
	ErrUserNotFound        = errors.NotFound("USER_NOT_FOUND", "用户不存在")
	ErrUserAlreadyExists   = errors.Conflict("USER_ALREADY_EXISTS", "用户名或邮箱已存在")
	ErrInvalidCredentials  = errors.Unauthorized("INVALID_CREDENTIALS", "账号或密码错误")
	ErrUserForbidden       = errors.Forbidden("USER_FORBIDDEN", "用户状态异常")
	ErrUnauthenticated     = errors.Unauthorized("UNAUTHENTICATED", "请先登录")
	ErrInvalidRefreshToken = errors.Unauthorized("INVALID_REFRESH_TOKEN", "刷新令牌无效")
)

type AuthRepo interface {
	SaveRefreshSession(ctx context.Context, refreshToken string, session *pkgAuth.RefreshSession, ttl time.Duration) error
	GetRefreshSession(ctx context.Context, refreshToken string) (*pkgAuth.RefreshSession, error)
	DeleteRefreshSession(ctx context.Context, refreshToken string) error
	DeleteRefreshSessionBySessionID(ctx context.Context, sessionID string) error
}

type UserRepo interface {
	Create(ctx context.Context, user *User) (*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByAccount(ctx context.Context, account string) (*User, error)
	UpdateProfile(ctx context.Context, user *User) (*User, error)
	UpdateLoginMeta(ctx context.Context, id int64, loginAt time.Time, loginIP string) error
	CreateStats(ctx context.Context, stats *UserStats) error
	GetStats(ctx context.Context, userID int64) (*UserStats, error)
}

type AuthUsecase struct {
	userRepo     UserRepo
	authRepo     AuthRepo
	tokenManager *pkgAuth.TokenManager
	log          *log.Helper
}

func NewAuthUsecase(userRepo UserRepo, authRepo AuthRepo, tokenManager *pkgAuth.TokenManager, logger log.Logger) *AuthUsecase {
	return &AuthUsecase{
		userRepo:     userRepo,
		authRepo:     authRepo,
		tokenManager: tokenManager,
		log:          log.NewHelper(logger),
	}
}

func (uc *AuthUsecase) Register(ctx context.Context, req *apiauth.RegisterRequest, loginIP string) (*User, string, string, int64, error) {
	username := strings.TrimSpace(req.Username)
	nickname := strings.TrimSpace(req.Nickname)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if nickname == "" {
		nickname = username
	}
	if err := pkgAuth.ValidateUsername(username); err != nil {
		return nil, "", "", 0, errors.BadRequest("INVALID_USERNAME", err.Error())
	}
	if err := pkgAuth.ValidatePassword(req.Password); err != nil {
		return nil, "", "", 0, errors.BadRequest("INVALID_PASSWORD", err.Error())
	}
	if _, err := uc.userRepo.FindByUsername(ctx, username); err == nil {
		return nil, "", "", 0, ErrUserAlreadyExists
	}
	if email != "" {
		if _, err := uc.userRepo.FindByEmail(ctx, email); err == nil {
			return nil, "", "", 0, ErrUserAlreadyExists
		}
	}
	hash, err := pkgAuth.HashPassword(req.Password)
	if err != nil {
		return nil, "", "", 0, err
	}
	now := time.Now()
	user, err := uc.userRepo.Create(ctx, &User{
		Username:     username,
		Nickname:     nickname,
		Email:        email,
		PasswordHash: hash,
		Role:         UserRoleNormal,
		Status:       UserStatusActive,
		LastLoginAt:  &now,
		LastLoginIP:  loginIP,
	})
	if err != nil {
		return nil, "", "", 0, err
	}
	if err := uc.userRepo.CreateStats(ctx, &UserStats{UserID: user.ID}); err != nil {
		return nil, "", "", 0, err
	}
	accessToken, refreshToken, expiresIn, err := uc.issueSession(ctx, user)
	if err != nil {
		return nil, "", "", 0, err
	}
	return user, accessToken, refreshToken, expiresIn, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, account, password, loginIP string) (*User, string, string, int64, error) {
	user, err := uc.userRepo.FindByAccount(ctx, strings.TrimSpace(account))
	if err != nil {
		return nil, "", "", 0, ErrInvalidCredentials
	}
	if user.Status == UserStatusBanned {
		return nil, "", "", 0, ErrUserForbidden
	}
	if err := pkgAuth.ComparePassword(user.PasswordHash, password); err != nil {
		return nil, "", "", 0, ErrInvalidCredentials
	}
	now := time.Now()
	if err := uc.userRepo.UpdateLoginMeta(ctx, user.ID, now, loginIP); err != nil {
		return nil, "", "", 0, err
	}
	user.LastLoginAt = &now
	user.LastLoginIP = loginIP
	accessToken, refreshToken, expiresIn, err := uc.issueSession(ctx, user)
	if err != nil {
		return nil, "", "", 0, err
	}
	return user, accessToken, refreshToken, expiresIn, nil
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (*User, string, string, int64, error) {
	session, err := uc.authRepo.GetRefreshSession(ctx, strings.TrimSpace(refreshToken))
	if err != nil || session == nil || session.ExpiresAt.Before(time.Now()) {
		return nil, "", "", 0, ErrInvalidRefreshToken
	}
	user, err := uc.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, "", "", 0, ErrUserNotFound
	}
	if user.Status == UserStatusBanned {
		return nil, "", "", 0, ErrUserForbidden
	}
	if err := uc.authRepo.DeleteRefreshSession(ctx, refreshToken); err != nil {
		return nil, "", "", 0, err
	}
	accessToken, newRefreshToken, expiresIn, err := uc.issueSession(ctx, user)
	if err != nil {
		return nil, "", "", 0, err
	}
	return user, accessToken, newRefreshToken, expiresIn, nil
}

func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken, sessionID string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken != "" {
		return uc.authRepo.DeleteRefreshSession(ctx, refreshToken)
	}
	if sessionID == "" {
		return nil
	}
	return uc.authRepo.DeleteRefreshSessionBySessionID(ctx, sessionID)
}

func (uc *AuthUsecase) GetCurrentUser(ctx context.Context, userID int64) (*User, error) {
	return uc.userRepo.FindByID(ctx, userID)
}

func (uc *AuthUsecase) issueSession(ctx context.Context, user *User) (string, string, int64, error) {
	sessionID := uuid.NewString()
	accessToken, expiresIn, err := uc.tokenManager.GenerateAccessToken(user.ID, user.Role, sessionID)
	if err != nil {
		return "", "", 0, err
	}
	refreshToken, err := uc.tokenManager.GenerateRefreshToken()
	if err != nil {
		return "", "", 0, err
	}
	expiresAt := time.Now().Add(uc.tokenManager.RefreshTokenTTL())
	session := &pkgAuth.RefreshSession{
		UserID:    user.ID,
		Role:      user.Role,
		SessionID: sessionID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if err := uc.authRepo.SaveRefreshSession(ctx, refreshToken, session, uc.tokenManager.RefreshTokenTTL()); err != nil {
		return "", "", 0, err
	}
	return accessToken, refreshToken, expiresIn, nil
}
