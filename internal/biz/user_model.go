package biz

import "time"

type User struct {
	ID           int64
	Username     string
	Nickname     string
	AvatarURL    string
	Bio          string
	Email        string
	PasswordHash string
	Role         int32
	Status       int32
	LastLoginAt  *time.Time
	LastLoginIP  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserStats struct {
	UserID                int64
	ArtworkCount          int64
	FollowerCount         int64
	FollowingCount        int64
	LikeReceivedCount     int64
	FavoriteReceivedCount int64
	CommentReceivedCount  int64
}
