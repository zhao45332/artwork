package service

import (
	"context"

	userv1 "artwork/api/user/v1"
	"artwork/internal/biz"
	"artwork/internal/pkg/xctx"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	userv1.UnimplementedUserServiceServer

	uc  *biz.UserUsecase
	log *log.Helper
}

func NewUserService(uc *biz.UserUsecase, logger log.Logger) *UserService {
	return &UserService{uc: uc, log: log.NewHelper(logger)}
}

func (s *UserService) GetUserProfile(ctx context.Context, req *userv1.GetUserProfileRequest) (*userv1.GetUserProfileReply, error) {
	user, stats, err := s.uc.GetUserProfile(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &userv1.GetUserProfileReply{User: s.toUserProfile(user, stats)}, nil
}

func (s *UserService) UpdateMyProfile(ctx context.Context, req *userv1.UpdateMyProfileRequest) (*userv1.UpdateMyProfileReply, error) {
	userID, ok := xctx.UserIDFromContext(ctx)
	if !ok {
		return nil, biz.ErrUnauthenticated
	}
	user, stats, err := s.uc.UpdateMyProfile(ctx, userID, req.Nickname, req.AvatarUrl, req.Bio)
	if err != nil {
		return nil, err
	}
	return &userv1.UpdateMyProfileReply{User: s.toUserProfile(user, stats)}, nil
}

func (s *UserService) ListCategories(ctx context.Context, _ *userv1.ListCategoriesRequest) (*userv1.ListCategoriesReply, error) {
	categories, err := s.uc.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*userv1.CategoryInfo, 0, len(categories))
	for _, category := range categories {
		items = append(items, &userv1.CategoryInfo{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
		})
	}
	return &userv1.ListCategoriesReply{Categories: items}, nil
}

func (s *UserService) toUserProfile(user *biz.User, stats *biz.UserStats) *userv1.UserProfile {
	profile := &userv1.UserProfile{
		Id:                    user.ID,
		Username:              user.Username,
		Nickname:              user.Nickname,
		AvatarUrl:             user.AvatarURL,
		Bio:                   user.Bio,
		Email:                 user.Email,
		Role:                  user.Role,
		Status:                user.Status,
		ArtworkCount:          stats.ArtworkCount,
		FollowerCount:         stats.FollowerCount,
		FollowingCount:        stats.FollowingCount,
		LikeReceivedCount:     stats.LikeReceivedCount,
		FavoriteReceivedCount: stats.FavoriteReceivedCount,
	}
	if !user.CreatedAt.IsZero() {
		profile.CreatedAt = timestamppb.New(user.CreatedAt)
	}
	if !user.UpdatedAt.IsZero() {
		profile.UpdatedAt = timestamppb.New(user.UpdatedAt)
	}
	return profile
}
