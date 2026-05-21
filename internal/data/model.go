package data

import (
	"time"

	"gorm.io/gorm"
)

const (
	ArtworkStatusDraft       int32 = 0
	ArtworkStatusPublished   int32 = 1
	ArtworkVisibilityPublic  int32 = 1
	ArtworkVisibilityPrivate int32 = 2
)

// Artwork 画作模型
type Artwork struct {
	ID            int64           `gorm:"primaryKey;autoIncrement"`
	UserID        int64           `gorm:"not null;index"`
	User          *User           `gorm:"foreignKey:UserID"`
	Title         string          `gorm:"type:varchar(200);not null;index"`
	Description   string          `gorm:"type:text"`
	Author        string          `gorm:"type:varchar(100);index"`
	Year          int32           `gorm:"index"`
	Width         float64         `gorm:"type:decimal(10,2)"`
	Height        float64         `gorm:"type:decimal(10,2)"`
	Material      string          `gorm:"type:varchar(100)"`
	Price         float64         `gorm:"type:decimal(12,2)"`
	IsCollected   bool            `gorm:"default:false"`
	Location      string          `gorm:"type:varchar(200)"`
	Style         string          `gorm:"type:varchar(100)"`
	Technique     string          `gorm:"type:varchar(100)"`
	ImageURL      string          `gorm:"type:varchar(500)"`
	CoverImageURL string          `gorm:"type:varchar(500)"`
	CategoryID    int64           `gorm:"index"`
	Category      *Category       `gorm:"foreignKey:CategoryID"`
	Tags          []*Tag          `gorm:"many2many:artwork_tags;"`
	Images        []*ArtworkImage `gorm:"foreignKey:ArtworkID"`
	Status        int32           `gorm:"default:1;index"`
	Visibility    int32           `gorm:"default:1;index"`
	ViewCount     int64           `gorm:"default:0"`
	LikeCount     int64           `gorm:"default:0"`
	FavoriteCount int64           `gorm:"default:0"`
	CommentCount  int64           `gorm:"default:0"`
	ShareCount    int64           `gorm:"default:0"`
	PublishAt     *time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Artwork) TableName() string {
	return "artworks"
}

type ArtworkImage struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	ArtworkID int64  `gorm:"not null;index"`
	URL       string `gorm:"type:varchar(500);not null"`
	IsCover   bool   `gorm:"default:false"`
	SortOrder int32  `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ArtworkImage) TableName() string {
	return "artwork_images"
}

// Category 分类模型
type Category struct {
	ID          int64      `gorm:"primaryKey;autoIncrement"`
	Name        string     `gorm:"type:varchar(100);not null;uniqueIndex"`
	Description string     `gorm:"type:text"`
	Artworks    []*Artwork `gorm:"foreignKey:CategoryID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Category) TableName() string {
	return "categories"
}

// Tag 标签模型
type Tag struct {
	ID         int64      `gorm:"primaryKey;autoIncrement"`
	Name       string     `gorm:"type:varchar(50);not null;uniqueIndex"`
	UsageCount int32      `gorm:"default:0"`
	Artworks   []*Artwork `gorm:"many2many:artwork_tags;"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (Tag) TableName() string {
	return "tags"
}

type ArtworkTag struct {
	ArtworkID int64 `gorm:"primaryKey"`
	TagID     int64 `gorm:"primaryKey"`
}

func (ArtworkTag) TableName() string {
	return "artwork_tags"
}
