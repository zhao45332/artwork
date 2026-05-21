package biz

import (
	"context"
	"strings"
	"time"

	"artwork/internal/pkg/xctx"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

const (
	ArtworkStatusDraft       int32 = 0
	ArtworkStatusPublished   int32 = 1
	ArtworkVisibilityPublic  int32 = 1
	ArtworkVisibilityPrivate int32 = 2
)

// Artwork 画作业务模型
type Artwork struct {
	ID            int64
	UserID        int64
	Username      string
	UserNickname  string
	UserAvatarURL string
	Title         string
	Description   string
	Author        string
	Year          int32
	Width         float64
	Height        float64
	Material      string
	Price         float64
	IsCollected   bool
	Location      string
	Style         string
	Technique     string
	ImageURL      string
	CoverImageURL string
	ImageURLs     []string
	CategoryID    int64
	Category      *Category
	Tags          []*Tag
	Status        int32
	Visibility    int32
	ViewCount     int64
	LikeCount     int64
	FavoriteCount int64
	CommentCount  int64
	ShareCount    int64
	PublishAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ArtworkImage struct {
	ID        int64
	URL       string
	IsCover   bool
	SortOrder int32
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
	Save(ctx context.Context, artwork *Artwork) (*Artwork, error)
	Update(ctx context.Context, artwork *Artwork) (*Artwork, error)
	Delete(ctx context.Context, id int64, actorUserID int64, actorRole int32) error
	FindByID(ctx context.Context, id int64, viewerUserID int64, viewerRole int32) (*Artwork, error)
	Search(ctx context.Context, req *SearchRequest) ([]*Artwork, int64, error)
}

// CategoryRepo 分类仓储接口
type CategoryRepo interface {
	FindByID(ctx context.Context, id int64) (*Category, error)
	List(ctx context.Context) ([]*Category, error)
}

// TagRepo 标签仓储接口
type TagRepo interface {
	FindOrCreateByName(ctx context.Context, name string) (*Tag, error)
	FindByNames(ctx context.Context, names []string) ([]*Tag, error)
}

// ImageRepo 图片仓储接口
type ImageRepo interface {
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
	UserID     int64
	ViewerID   int64
	ViewerRole int32
	Visibility int32
	Status     int32
}

var (
	ErrArtworkNotFound  = kratosErrors.NotFound("ARTWORK_NOT_FOUND", "画作不存在")
	ErrCategoryNotFound = kratosErrors.NotFound("CATEGORY_NOT_FOUND", "分类不存在")
	ErrInvalidParams    = kratosErrors.BadRequest("INVALID_PARAMS", "参数无效")
	ErrPermissionDenied = kratosErrors.Forbidden("PERMISSION_DENIED", "无权限执行该操作")
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

func (uc *ArtworkUsecase) CreateArtwork(ctx context.Context, artwork *Artwork) (*Artwork, error) {
	userID, ok := xctx.UserIDFromContext(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}
	artwork.UserID = userID
	if artwork.Visibility == 0 {
		artwork.Visibility = ArtworkVisibilityPublic
	}
	if artwork.Status != ArtworkStatusDraft {
		artwork.Status = ArtworkStatusPublished
		now := time.Now()
		artwork.PublishAt = &now
	}
	if err := uc.normalizeArtwork(ctx, artwork); err != nil {
		return nil, err
	}
	uc.log.Infof("正在创建画作: %s", artwork.Title)
	return uc.artworkRepo.Save(ctx, artwork)
}

func (uc *ArtworkUsecase) UpdateArtwork(ctx context.Context, artwork *Artwork) (*Artwork, error) {
	userID, _ := xctx.UserIDFromContext(ctx)
	role := currentRole(ctx)
	existing, err := uc.artworkRepo.FindByID(ctx, artwork.ID, userID, role)
	if err != nil {
		return nil, err
	}
	if existing.UserID != userID && role != UserRoleAdmin {
		return nil, ErrPermissionDenied
	}
	artwork.UserID = existing.UserID
	if artwork.Visibility == 0 {
		artwork.Visibility = existing.Visibility
	}
	if artwork.Status == 0 && existing.Status != ArtworkStatusDraft {
		artwork.Status = existing.Status
	}
	if artwork.Status == ArtworkStatusPublished {
		now := time.Now()
		artwork.PublishAt = &now
	}
	if err := uc.normalizeArtwork(ctx, artwork); err != nil {
		return nil, err
	}
	uc.log.Infof("正在更新画作: %d", artwork.ID)
	return uc.artworkRepo.Update(ctx, artwork)
}

func (uc *ArtworkUsecase) DeleteArtwork(ctx context.Context, id int64) error {
	userID, _ := xctx.UserIDFromContext(ctx)
	role := currentRole(ctx)
	uc.log.Infof("正在删除画作: %d", id)
	return uc.artworkRepo.Delete(ctx, id, userID, role)
}

func (uc *ArtworkUsecase) GetArtwork(ctx context.Context, id int64) (*Artwork, error) {
	userID, _ := xctx.UserIDFromContext(ctx)
	return uc.artworkRepo.FindByID(ctx, id, userID, currentRole(ctx))
}

func (uc *ArtworkUsecase) SearchArtwork(ctx context.Context, req *SearchRequest) ([]*Artwork, int64, error) {
	viewerID, _ := xctx.UserIDFromContext(ctx)
	req.ViewerID = viewerID
	req.ViewerRole = currentRole(ctx)
	if req.Status == 0 {
		req.Status = ArtworkStatusPublished
	}
	if req.Visibility == 0 {
		req.Visibility = ArtworkVisibilityPublic
	}
	return uc.artworkRepo.Search(ctx, req)
}

func (uc *ArtworkUsecase) GetMyArtworks(ctx context.Context, page, pageSize, status int32) ([]*Artwork, int64, error) {
	userID, ok := xctx.UserIDFromContext(ctx)
	if !ok {
		return nil, 0, ErrUnauthenticated
	}
	return uc.artworkRepo.Search(ctx, &SearchRequest{
		UserID:     userID,
		ViewerID:   userID,
		ViewerRole: currentRole(ctx),
		Page:       page,
		PageSize:   pageSize,
		Status:     status,
	})
}

func (uc *ArtworkUsecase) ListCategories(ctx context.Context) ([]*Category, error) {
	return uc.categoryRepo.List(ctx)
}

func (uc *ArtworkUsecase) UploadImage(ctx context.Context, filename string, content []byte, contentType string) (string, error) {
	if _, ok := xctx.UserIDFromContext(ctx); !ok {
		return "", ErrUnauthenticated
	}
	uc.log.Infof("正在上传图片: %s, 大小: %d 字节", filename, len(content))
	return uc.imageRepo.UploadImage(ctx, filename, content, contentType, uc.bucketName)
}

func (uc *ArtworkUsecase) normalizeArtwork(ctx context.Context, artwork *Artwork) error {
	artwork.Title = strings.TrimSpace(artwork.Title)
	artwork.Author = strings.TrimSpace(artwork.Author)
	if artwork.Title == "" {
		return ErrInvalidParams
	}
	if artwork.CategoryID > 0 {
		if _, err := uc.categoryRepo.FindByID(ctx, artwork.CategoryID); err != nil {
			return ErrCategoryNotFound
		}
	}
	if artwork.CoverImageURL == "" && artwork.ImageURL != "" {
		artwork.CoverImageURL = artwork.ImageURL
	}
	if artwork.CoverImageURL == "" && len(artwork.ImageURLs) > 0 {
		artwork.CoverImageURL = artwork.ImageURLs[0]
	}
	if artwork.ImageURL == "" && artwork.CoverImageURL != "" {
		artwork.ImageURL = artwork.CoverImageURL
	}
	filteredImages := make([]string, 0, len(artwork.ImageURLs))
	seen := map[string]struct{}{}
	for _, url := range artwork.ImageURLs {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		filteredImages = append(filteredImages, url)
	}
	if len(filteredImages) == 0 && artwork.ImageURL != "" {
		filteredImages = append(filteredImages, artwork.ImageURL)
	}
	artwork.ImageURLs = filteredImages
	if len(artwork.Tags) > 0 {
		tagNames := make([]string, 0, len(artwork.Tags))
		for _, tag := range artwork.Tags {
			name := strings.TrimSpace(tag.Name)
			if name != "" {
				tagNames = append(tagNames, name)
			}
		}
		tags, err := uc.tagRepo.FindByNames(ctx, tagNames)
		if err != nil {
			return err
		}
		artwork.Tags = tags
	}
	return nil
}

func currentRole(ctx context.Context) int32 {
	if authUser, ok := xctx.AuthUserFromContext(ctx); ok {
		return authUser.Role
	}
	return 0
}
