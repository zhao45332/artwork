package data

import "time"

type User struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	Username     string `gorm:"type:varchar(64);not null;uniqueIndex"`
	Nickname     string `gorm:"type:varchar(64);not null"`
	AvatarURL    string `gorm:"type:varchar(500)"`
	Bio          string `gorm:"type:varchar(255)"`
	Email        string `gorm:"type:varchar(128);uniqueIndex"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	Role         int32  `gorm:"default:1"`
	Status       int32  `gorm:"default:1"`
	LastLoginAt  *time.Time
	LastLoginIP  string `gorm:"type:varchar(64)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string {
	return "users"
}

type UserStats struct {
	UserID                int64 `gorm:"primaryKey"`
	ArtworkCount          int64 `gorm:"default:0"`
	FollowerCount         int64 `gorm:"default:0"`
	FollowingCount        int64 `gorm:"default:0"`
	LikeReceivedCount     int64 `gorm:"default:0"`
	FavoriteReceivedCount int64 `gorm:"default:0"`
	CommentReceivedCount  int64 `gorm:"default:0"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (UserStats) TableName() string {
	return "user_stats"
}

type UserIdentity struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	UserID       int64  `gorm:"not null;index"`
	IdentityType string `gorm:"type:varchar(32);not null"`
	IdentityKey  string `gorm:"type:varchar(128);not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserIdentity) TableName() string {
	return "user_identities"
}
