package biz

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type UserUsecase struct {
	userRepo     UserRepo
	categoryRepo CategoryRepo
	log          *log.Helper
}

func NewUserUsecase(userRepo UserRepo, categoryRepo CategoryRepo, logger log.Logger) *UserUsecase {
	return &UserUsecase{
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		log:          log.NewHelper(logger),
	}
}

func (uc *UserUsecase) GetUserProfile(ctx context.Context, userID int64) (*User, *UserStats, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}
	stats, err := uc.userRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return user, stats, nil
}

func (uc *UserUsecase) UpdateMyProfile(ctx context.Context, userID int64, nickname, avatarURL, bio string) (*User, *UserStats, error) {
	if userID == 0 {
		return nil, nil, ErrUnauthenticated
	}
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return nil, nil, errors.BadRequest("INVALID_NICKNAME", "昵称不能为空")
	}
	updated, err := uc.userRepo.UpdateProfile(ctx, &User{
		ID:        userID,
		Nickname:  nickname,
		AvatarURL: strings.TrimSpace(avatarURL),
		Bio:       strings.TrimSpace(bio),
	})
	if err != nil {
		return nil, nil, err
	}
	stats, err := uc.userRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return updated, stats, nil
}

func (uc *UserUsecase) ListCategories(ctx context.Context) ([]*Category, error) {
	return uc.categoryRepo.List(ctx)
}
