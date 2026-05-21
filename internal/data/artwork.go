package data

import (
	"context"
	"fmt"
	"strings"

	"artwork/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type artworkRepo struct {
	data *Data
	log  *log.Helper
}

func NewArtworkRepo(data *Data, logger log.Logger) biz.ArtworkRepo {
	return &artworkRepo{data: data, log: log.NewHelper(logger)}
}

func (r *artworkRepo) Save(ctx context.Context, artwork *biz.Artwork) (*biz.Artwork, error) {
	artworkModel := &Artwork{
		UserID:        artwork.UserID,
		Title:         artwork.Title,
		Description:   artwork.Description,
		Author:        artwork.Author,
		Year:          artwork.Year,
		Width:         artwork.Width,
		Height:        artwork.Height,
		Material:      artwork.Material,
		Price:         artwork.Price,
		IsCollected:   artwork.IsCollected,
		Location:      artwork.Location,
		Style:         artwork.Style,
		Technique:     artwork.Technique,
		ImageURL:      artwork.ImageURL,
		CoverImageURL: artwork.CoverImageURL,
		CategoryID:    artwork.CategoryID,
		Status:        artwork.Status,
		Visibility:    artwork.Visibility,
		PublishAt:     artwork.PublishAt,
	}
	if len(artwork.Tags) > 0 {
		tagModels := make([]*Tag, 0, len(artwork.Tags))
		for _, tag := range artwork.Tags {
			tagModels = append(tagModels, &Tag{ID: tag.ID})
		}
		artworkModel.Tags = tagModels
	}
	for idx, url := range artwork.ImageURLs {
		artworkModel.Images = append(artworkModel.Images, &ArtworkImage{
			URL:       url,
			IsCover:   url == artwork.CoverImageURL,
			SortOrder: int32(idx),
		})
	}
	if err := r.data.db.WithContext(ctx).Create(artworkModel).Error; err != nil {
		r.log.Errorf("创建画作失败: %v", err)
		return nil, err
	}
	return r.FindByID(ctx, artworkModel.ID, artwork.UserID, biz.UserRoleAdmin)
}

func (r *artworkRepo) Update(ctx context.Context, artwork *biz.Artwork) (*biz.Artwork, error) {
	var result *biz.Artwork
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		artworkModel := &Artwork{}
		if err := tx.Preload("Images").First(artworkModel, artwork.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return biz.ErrArtworkNotFound
			}
			return err
		}
		artworkModel.Title = artwork.Title
		artworkModel.Description = artwork.Description
		artworkModel.Author = artwork.Author
		artworkModel.Year = artwork.Year
		artworkModel.Width = artwork.Width
		artworkModel.Height = artwork.Height
		artworkModel.Material = artwork.Material
		artworkModel.Price = artwork.Price
		artworkModel.IsCollected = artwork.IsCollected
		artworkModel.Location = artwork.Location
		artworkModel.Style = artwork.Style
		artworkModel.Technique = artwork.Technique
		artworkModel.ImageURL = artwork.ImageURL
		artworkModel.CoverImageURL = artwork.CoverImageURL
		artworkModel.CategoryID = artwork.CategoryID
		artworkModel.Status = artwork.Status
		artworkModel.Visibility = artwork.Visibility
		artworkModel.PublishAt = artwork.PublishAt

		if err := tx.Model(artworkModel).Association("Tags").Clear(); err != nil {
			return err
		}
		if len(artwork.Tags) > 0 {
			tagModels := make([]*Tag, 0, len(artwork.Tags))
			for _, tag := range artwork.Tags {
				tagModels = append(tagModels, &Tag{ID: tag.ID})
			}
			if err := tx.Model(artworkModel).Association("Tags").Replace(tagModels); err != nil {
				return err
			}
		}
		if err := tx.Where("artwork_id = ?", artwork.ID).Delete(&ArtworkImage{}).Error; err != nil {
			return err
		}
		for idx, url := range artwork.ImageURLs {
			if err := tx.Create(&ArtworkImage{
				ArtworkID: artwork.ID,
				URL:       url,
				IsCover:   url == artwork.CoverImageURL,
				SortOrder: int32(idx),
			}).Error; err != nil {
				return err
			}
		}
		if err := tx.Save(artworkModel).Error; err != nil {
			return err
		}
		var err error
		result, err = r.findByID(tx.WithContext(ctx), artwork.ID, artwork.UserID, biz.UserRoleAdmin)
		return err
	})
	if err != nil {
		r.log.Errorf("更新画作失败: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *artworkRepo) Delete(ctx context.Context, id int64, actorUserID int64, actorRole int32) error {
	artwork, err := r.findByID(r.data.db.WithContext(ctx), id, actorUserID, actorRole)
	if err != nil {
		return err
	}
	if artwork.UserID != actorUserID && actorRole != biz.UserRoleAdmin {
		return biz.ErrPermissionDenied
	}
	return r.data.db.WithContext(ctx).Delete(&Artwork{}, id).Error
}

func (r *artworkRepo) FindByID(ctx context.Context, id int64, viewerUserID int64, viewerRole int32) (*biz.Artwork, error) {
	return r.findByID(r.data.db.WithContext(ctx), id, viewerUserID, viewerRole)
}

func (r *artworkRepo) findByID(db *gorm.DB, id int64, viewerUserID int64, viewerRole int32) (*biz.Artwork, error) {
	artworkModel := &Artwork{}
	if err := db.Preload("User").Preload("Category").Preload("Tags").Preload("Images", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order ASC")
	}).First(artworkModel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrArtworkNotFound
		}
		return nil, err
	}
	if artworkModel.Visibility == ArtworkVisibilityPrivate && artworkModel.UserID != viewerUserID && viewerRole != biz.UserRoleAdmin {
		return nil, biz.ErrArtworkNotFound
	}
	if artworkModel.Status != ArtworkStatusPublished && artworkModel.UserID != viewerUserID && viewerRole != biz.UserRoleAdmin {
		return nil, biz.ErrArtworkNotFound
	}
	if artworkModel.UserID != viewerUserID && viewerRole != biz.UserRoleAdmin {
		db.Model(&Artwork{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
		artworkModel.ViewCount++
	}
	return r.toBizArtwork(artworkModel), nil
}

func (r *artworkRepo) Search(ctx context.Context, req *biz.SearchRequest) ([]*biz.Artwork, int64, error) {
	query := r.data.db.WithContext(ctx).Model(&Artwork{}).
		Preload("User").
		Preload("Category").
		Preload("Tags").
		Preload("Images", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") })

	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
		if req.UserID != req.ViewerID && req.ViewerRole != biz.UserRoleAdmin {
			query = query.Where("status = ? AND visibility = ?", ArtworkStatusPublished, ArtworkVisibilityPublic)
		}
	} else {
		query = query.Where("status = ? AND visibility = ?", ArtworkStatusPublished, ArtworkVisibilityPublic)
		if req.ViewerID > 0 {
			query = query.Or("user_id = ?", req.ViewerID)
		}
	}
	if req.Status >= 0 && req.UserID == req.ViewerID && req.UserID > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR author LIKE ?", keyword, keyword, keyword)
	}
	if req.YearFrom > 0 {
		query = query.Where("year >= ?", req.YearFrom)
	}
	if req.YearTo > 0 {
		query = query.Where("year <= ?", req.YearTo)
	}
	if req.WidthMin > 0 {
		query = query.Where("width >= ?", req.WidthMin)
	}
	if req.WidthMax > 0 {
		query = query.Where("width <= ?", req.WidthMax)
	}
	if req.HeightMin > 0 {
		query = query.Where("height >= ?", req.HeightMin)
	}
	if req.HeightMax > 0 {
		query = query.Where("height <= ?", req.HeightMax)
	}
	if req.PriceMin > 0 {
		query = query.Where("price >= ?", req.PriceMin)
	}
	if req.PriceMax > 0 {
		query = query.Where("price <= ?", req.PriceMax)
	}
	if req.CategoryID > 0 {
		query = query.Where("category_id = ?", req.CategoryID)
	}
	if len(req.Tags) > 0 {
		query = query.Joins("JOIN artwork_tags ON artworks.id = artwork_tags.artwork_id").
			Joins("JOIN tags ON artwork_tags.tag_id = tags.id").
			Where("tags.name IN ?", req.Tags).
			Group("artworks.id")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	sortBy := sanitizeSortBy(req.SortBy)
	sortOrder := strings.ToUpper(req.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 12
	}
	if pageSize > 60 {
		pageSize = 60
	}
	var artworkModels []*Artwork
	if err := query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&artworkModels).Error; err != nil {
		return nil, 0, err
	}
	artworks := make([]*biz.Artwork, 0, len(artworkModels))
	for _, model := range artworkModels {
		artworks = append(artworks, r.toBizArtwork(model))
	}
	return artworks, total, nil
}

func sanitizeSortBy(sortBy string) string {
	switch sortBy {
	case "created_at", "updated_at", "title", "price", "view_count", "like_count", "favorite_count":
		return sortBy
	default:
		return "created_at"
	}
}

func (r *artworkRepo) toBizArtwork(model *Artwork) *biz.Artwork {
	artwork := &biz.Artwork{
		ID:            model.ID,
		UserID:        model.UserID,
		Title:         model.Title,
		Description:   model.Description,
		Author:        model.Author,
		Year:          model.Year,
		Width:         model.Width,
		Height:        model.Height,
		Material:      model.Material,
		Price:         model.Price,
		IsCollected:   model.IsCollected,
		Location:      model.Location,
		Style:         model.Style,
		Technique:     model.Technique,
		ImageURL:      model.ImageURL,
		CoverImageURL: model.CoverImageURL,
		CategoryID:    model.CategoryID,
		Status:        model.Status,
		Visibility:    model.Visibility,
		ViewCount:     model.ViewCount,
		LikeCount:     model.LikeCount,
		FavoriteCount: model.FavoriteCount,
		CommentCount:  model.CommentCount,
		ShareCount:    model.ShareCount,
		PublishAt:     model.PublishAt,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
	if model.User != nil {
		artwork.Username = model.User.Username
		artwork.UserNickname = model.User.Nickname
		artwork.UserAvatarURL = model.User.AvatarURL
	}
	if model.Category != nil {
		artwork.Category = &biz.Category{ID: model.Category.ID, Name: model.Category.Name, Description: model.Category.Description}
	}
	if len(model.Tags) > 0 {
		artwork.Tags = make([]*biz.Tag, 0, len(model.Tags))
		for _, tag := range model.Tags {
			artwork.Tags = append(artwork.Tags, &biz.Tag{ID: tag.ID, Name: tag.Name, UsageCount: tag.UsageCount})
		}
	}
	if len(model.Images) > 0 {
		artwork.ImageURLs = make([]string, 0, len(model.Images))
		for _, img := range model.Images {
			artwork.ImageURLs = append(artwork.ImageURLs, img.URL)
			if img.IsCover && artwork.CoverImageURL == "" {
				artwork.CoverImageURL = img.URL
			}
		}
	}
	if artwork.CoverImageURL == "" {
		artwork.CoverImageURL = artwork.ImageURL
	}
	return artwork
}
