package data

import (
	"context"
	"errors"
	"strings"
	"time"

	"artwork/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type UserDataRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewUserDataRepo(db *gorm.DB, logger log.Logger) *UserDataRepo {
	return &UserDataRepo{
		db:  db,
		log: log.NewHelper(logger),
	}
}

func (r *UserDataRepo) Create(ctx context.Context, user *biz.User) (*biz.User, error) {
	dataUser := &User{
		Username:     user.Username,
		Nickname:     user.Nickname,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
		Status:       user.Status,
	}
	if user.LastLoginAt != nil {
		dataUser.LastLoginAt = user.LastLoginAt
	}
	if user.LastLoginIP != "" {
		dataUser.LastLoginIP = user.LastLoginIP
	}
	if err := r.db.WithContext(ctx).Create(dataUser).Error; err != nil {
		return nil, err
	}
	return r.toBizUser(dataUser), nil
}

func (r *UserDataRepo) FindByID(ctx context.Context, id int64) (*biz.User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return r.toBizUser(&user), nil
}

func (r *UserDataRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return r.toBizUser(&user), nil
}

func (r *UserDataRepo) FindByEmail(ctx context.Context, email string) (*biz.User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return r.toBizUser(&user), nil
}

func (r *UserDataRepo) FindByAccount(ctx context.Context, account string) (*biz.User, error) {
	account = strings.TrimSpace(account)
	var user User
	err := r.db.WithContext(ctx).
		Where("username = ? OR email = ?", account, account).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return r.toBizUser(&user), nil
}

func (r *UserDataRepo) UpdateProfile(ctx context.Context, user *biz.User) (*biz.User, error) {
	updates := map[string]any{
		"nickname":   user.Nickname,
		"avatar_url": user.AvatarURL,
		"bio":        user.Bio,
	}
	if err := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.FindByID(ctx, user.ID)
}

func (r *UserDataRepo) UpdateLoginMeta(ctx context.Context, id int64, loginAt time.Time, loginIP string) error {
	return r.db.WithContext(ctx).Model(&User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_login_at": loginAt,
			"last_login_ip": loginIP,
		}).Error
}

func (r *UserDataRepo) CreateStats(ctx context.Context, stats *biz.UserStats) error {
	return r.db.WithContext(ctx).Create(&UserStats{UserID: stats.UserID}).Error
}

func (r *UserDataRepo) GetStats(ctx context.Context, userID int64) (*biz.UserStats, error) {
	var stats UserStats
	if err := r.db.WithContext(ctx).First(&stats, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &biz.UserStats{UserID: userID}, nil
		}
		return nil, err
	}
	return r.toBizStats(&stats), nil
}

func (r *UserDataRepo) toBizUser(u *User) *biz.User {
	return &biz.User{
		ID:           u.ID,
		Username:     u.Username,
		Nickname:     u.Nickname,
		AvatarURL:    u.AvatarURL,
		Bio:          u.Bio,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		Status:       u.Status,
		LastLoginAt:  u.LastLoginAt,
		LastLoginIP:  u.LastLoginIP,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func (r *UserDataRepo) toBizStats(s *UserStats) *biz.UserStats {
	return &biz.UserStats{
		UserID:                s.UserID,
		ArtworkCount:          s.ArtworkCount,
		FollowerCount:         s.FollowerCount,
		FollowingCount:        s.FollowingCount,
		LikeReceivedCount:     s.LikeReceivedCount,
		FavoriteReceivedCount: s.FavoriteReceivedCount,
		CommentReceivedCount:  s.CommentReceivedCount,
	}
}
