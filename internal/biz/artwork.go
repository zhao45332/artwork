package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// Artwork 画作业务模型
type Artwork struct {
	ID          int64
	Title       string
	Description string
	Author      string
	Year        int32
	Width       float64
	Height      float64
	Material    string
	Price       float64
	IsCollected bool
	Location    string
	Style       string
	Technique   string
	ImageURL    string
	CategoryID  int64
	Category    *Category
	Tags        []*Tag
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Category 分类业务模型
type Category struct {
	ID          int64
	Name        string
	Description string
}

// Tag 标签业务模型
type Tag struct {
	ID         int64
	Name       string
	UsageCount int32
}

// ArtworkRepo 画作仓储接口
type ArtworkRepo interface {
	// Save 创建画作
	Save(ctx context.Context, artwork *Artwork) (*Artwork, error)
	// Update 更新画作
	Update(ctx context.Context, artwork *Artwork) (*Artwork, error)
	// Delete 软删除画作
	Delete(ctx context.Context, id int64) error
	// FindByID 根据ID查找画作
	FindByID(ctx context.Context, id int64) (*Artwork, error)
	// Search 搜索画作
	Search(ctx context.Context, req *SearchRequest) ([]*Artwork, int64, error)
}

// CategoryRepo 分类仓储接口
type CategoryRepo interface {
	// FindByID 根据ID查找分类
	FindByID(ctx context.Context, id int64) (*Category, error)
}

// TagRepo 标签仓储接口
type TagRepo interface {
	// FindOrCreateByName 根据名称查找或创建标签
	FindOrCreateByName(ctx context.Context, name string) (*Tag, error)
	// FindByNames 根据名称列表查找标签
	FindByNames(ctx context.Context, names []string) ([]*Tag, error)
}

// ImageRepo 图片仓储接口
type ImageRepo interface {
	// UploadImage 上传图片
	UploadImage(ctx context.Context, filename string, content []byte, contentType string, bucketName string) (string, error)
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword    string
	YearFrom   int32
	YearTo     int32
	WidthMin   float64
	WidthMax   float64
	HeightMin  float64
	HeightMax  float64
	PriceMin   float64
	PriceMax   float64
	CategoryID int64
	Tags       []string
	Page       int32
	PageSize   int32
	SortBy     string
	SortOrder  string
}

var (
	// ErrArtworkNotFound 画作不存在
	ErrArtworkNotFound = errors.NotFound("ARTWORK_NOT_FOUND", "画作不存在")
	// ErrCategoryNotFound 分类不存在
	ErrCategoryNotFound = errors.NotFound("CATEGORY_NOT_FOUND", "分类不存在")
	// ErrInvalidParams 参数无效
	ErrInvalidParams = errors.BadRequest("INVALID_PARAMS", "参数无效")
)

// ArtworkUsecase 画作用例
type ArtworkUsecase struct {
	artworkRepo  ArtworkRepo
	categoryRepo CategoryRepo
	tagRepo      TagRepo
	imageRepo    ImageRepo
	log          *log.Helper
	bucketName   string
}

// NewArtworkUsecase 创建画作用例
func NewArtworkUsecase(artworkRepo ArtworkRepo, categoryRepo CategoryRepo, tagRepo TagRepo, imageRepo ImageRepo, bucketName string, logger log.Logger) *ArtworkUsecase {
	return &ArtworkUsecase{
		artworkRepo:  artworkRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		imageRepo:    imageRepo,
		bucketName:   bucketName,
		log:          log.NewHelper(logger),
	}
}

// CreateArtwork 创建画作
func (uc *ArtworkUsecase) CreateArtwork(ctx context.Context, artwork *Artwork) (*Artwork, error) {
	// 验证分类是否存在
	if artwork.CategoryID > 0 {
		_, err := uc.categoryRepo.FindByID(ctx, artwork.CategoryID)
		if err != nil {
			uc.log.Errorf("分类不存在: %v", err)
			return nil, ErrCategoryNotFound
		}
	}

	// 处理标签
	if len(artwork.Tags) > 0 {
		tagNames := make([]string, 0, len(artwork.Tags))
		for _, tag := range artwork.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		tags, err := uc.tagRepo.FindByNames(ctx, tagNames)
		if err != nil {
			uc.log.Errorf("查找标签失败: %v", err)
			return nil, err
		}
		artwork.Tags = tags
	}

	uc.log.Infof("正在创建画作: %s", artwork.Title)
	return uc.artworkRepo.Save(ctx, artwork)
}

// UpdateArtwork 更新画作
func (uc *ArtworkUsecase) UpdateArtwork(ctx context.Context, artwork *Artwork) (*Artwork, error) {
	// 检查画作是否存在
	existing, err := uc.artworkRepo.FindByID(ctx, artwork.ID)
	if err != nil {
		return nil, err
	}

	// 验证分类是否存在
	if artwork.CategoryID > 0 && artwork.CategoryID != existing.CategoryID {
		_, err := uc.categoryRepo.FindByID(ctx, artwork.CategoryID)
		if err != nil {
			uc.log.Errorf("分类不存在: %v", err)
			return nil, ErrCategoryNotFound
		}
	}

	// 处理标签
	if len(artwork.Tags) > 0 {
		tagNames := make([]string, 0, len(artwork.Tags))
		for _, tag := range artwork.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		tags, err := uc.tagRepo.FindByNames(ctx, tagNames)
		if err != nil {
			uc.log.Errorf("查找标签失败: %v", err)
			return nil, err
		}
		artwork.Tags = tags
	}

	uc.log.Infof("正在更新画作: %d", artwork.ID)
	return uc.artworkRepo.Update(ctx, artwork)
}

// DeleteArtwork 删除画作
func (uc *ArtworkUsecase) DeleteArtwork(ctx context.Context, id int64) error {
	uc.log.Infof("正在删除画作: %d", id)
	return uc.artworkRepo.Delete(ctx, id)
}

// GetArtwork 获取画作
func (uc *ArtworkUsecase) GetArtwork(ctx context.Context, id int64) (*Artwork, error) {
	return uc.artworkRepo.FindByID(ctx, id)
}

// SearchArtwork 搜索画作
func (uc *ArtworkUsecase) SearchArtwork(ctx context.Context, req *SearchRequest) ([]*Artwork, int64, error) {
	uc.log.Infof("正在搜索画作，关键词: %s", req.Keyword)
	return uc.artworkRepo.Search(ctx, req)
}

// UploadImage 上传图片
func (uc *ArtworkUsecase) UploadImage(ctx context.Context, filename string, content []byte, contentType string) (string, error) {
	uc.log.Infof("正在上传图片: %s, 大小: %d 字节", filename, len(content))
	return uc.imageRepo.UploadImage(ctx, filename, content, contentType, uc.bucketName)
}
