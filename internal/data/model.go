package data

import (
	"time"

	"gorm.io/gorm"
)

// Artwork 画作模型
type Artwork struct {
	ID          int64          `gorm:"primaryKey;autoIncrement"`
	Title       string         `gorm:"type:varchar(200);not null;index"` // 标题
	Description string         `gorm:"type:text"`                        // 描述
	Author      string         `gorm:"type:varchar(100);index"`          // 作者
	Year        int32          `gorm:"index"`                            // 创作年份
	Width       float64        `gorm:"type:decimal(10,2)"`               // 宽度（cm）
	Height      float64        `gorm:"type:decimal(10,2)"`               // 高度（cm）
	Material    string         `gorm:"type:varchar(100)"`                // 材质
	Price       float64        `gorm:"type:decimal(12,2)"`               // 价格
	IsCollected bool           `gorm:"default:false"`                    // 收藏状态
	Location    string         `gorm:"type:varchar(200)"`                // 创作地点
	Style       string         `gorm:"type:varchar(100)"`                // 风格
	Technique   string         `gorm:"type:varchar(100)"`                // 技法
	ImageURL    string         `gorm:"type:varchar(500)"`                // 图片URL
	CategoryID  int64          `gorm:"index"`                            // 分类ID（一对多）
	Category    *Category      `gorm:"foreignKey:CategoryID"`            // 分类关联
	Tags        []*Tag         `gorm:"many2many:artwork_tags;"`          // 标签关联（多对多）
	DeletedAt   gorm.DeletedAt `gorm:"index"`                            // 软删除
	CreatedAt   time.Time      // 创建时间
	UpdatedAt   time.Time      // 更新时间
}

// TableName 指定表名
func (Artwork) TableName() string {
	return "artworks"
}

// Category 分类模型
type Category struct {
	ID          int64      `gorm:"primaryKey;autoIncrement"`
	Name        string     `gorm:"type:varchar(100);not null;uniqueIndex"` // 分类名称
	Description string     `gorm:"type:text"`                              // 分类描述
	Artworks    []*Artwork `gorm:"foreignKey:CategoryID"`                  // 画作关联
	CreatedAt   time.Time  // 创建时间
	UpdatedAt   time.Time  // 更新时间
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}

// Tag 标签模型
type Tag struct {
	ID         int64      `gorm:"primaryKey;autoIncrement"`
	Name       string     `gorm:"type:varchar(50);not null;uniqueIndex"` // 标签名称
	UsageCount int32      `gorm:"default:0"`                             // 使用次数
	Artworks   []*Artwork `gorm:"many2many:artwork_tags;"`               // 画作关联（多对多）
	CreatedAt  time.Time  // 创建时间
	UpdatedAt  time.Time  // 更新时间
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}

// ArtworkTag 画作标签关联表（GORM会自动创建，这里定义用于查询）
type ArtworkTag struct {
	ArtworkID int64 `gorm:"primaryKey"`
	TagID     int64 `gorm:"primaryKey"`
}

// TableName 指定表名
func (ArtworkTag) TableName() string {
	return "artwork_tags"
}
