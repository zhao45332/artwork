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

// NewArtworkRepo 创建画作仓储
func NewArtworkRepo(data *Data, logger log.Logger) biz.ArtworkRepo {
	return &artworkRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Save 创建画作
func (r *artworkRepo) Save(ctx context.Context, artwork *biz.Artwork) (*biz.Artwork, error) {
	// 转换为数据模型
	artworkModel := &Artwork{
		Title:       artwork.Title,
		Description: artwork.Description,
		Author:      artwork.Author,
		Year:        artwork.Year,
		Width:       artwork.Width,
		Height:      artwork.Height,
		Material:    artwork.Material,
		Price:       artwork.Price,
		IsCollected: artwork.IsCollected,
		Location:    artwork.Location,
		Style:       artwork.Style,
		Technique:   artwork.Technique,
		ImageURL:    artwork.ImageURL,
		CategoryID:  artwork.CategoryID,
	}

	// 处理标签关联
	if len(artwork.Tags) > 0 {
		tagModels := make([]*Tag, 0, len(artwork.Tags))
		for _, tag := range artwork.Tags {
			tagModels = append(tagModels, &Tag{ID: tag.ID})
		}
		artworkModel.Tags = tagModels
	}

	if err := r.data.db.WithContext(ctx).Create(artworkModel).Error; err != nil {
		r.log.Errorf("创建画作失败: %v", err)
		return nil, err
	}

	r.log.Infof("画作创建成功: ID=%d, Title=%s", artworkModel.ID, artworkModel.Title)
	return r.toBizArtwork(artworkModel), nil
}

// Update 更新画作
func (r *artworkRepo) Update(ctx context.Context, artwork *biz.Artwork) (*biz.Artwork, error) {
	artworkModel := &Artwork{}
	if err := r.data.db.WithContext(ctx).First(artworkModel, artwork.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrArtworkNotFound
		}
		r.log.Errorf("查找画作失败: %v", err)
		return nil, err
	}

	// 更新字段
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
	artworkModel.CategoryID = artwork.CategoryID

	// 更新标签关联
	if len(artwork.Tags) > 0 {
		tagModels := make([]*Tag, 0, len(artwork.Tags))
		for _, tag := range artwork.Tags {
			tagModels = append(tagModels, &Tag{ID: tag.ID})
		}
		// 替换关联
		if err := r.data.db.WithContext(ctx).Model(artworkModel).Association("Tags").Replace(tagModels); err != nil {
			r.log.Errorf("更新标签失败: %v", err)
			return nil, err
		}
	}

	if err := r.data.db.WithContext(ctx).Save(artworkModel).Error; err != nil {
		r.log.Errorf("更新画作失败: %v", err)
		return nil, err
	}

	r.log.Infof("画作更新成功: ID=%d", artworkModel.ID)
	return r.toBizArtwork(artworkModel), nil
}

// Delete 软删除画作
func (r *artworkRepo) Delete(ctx context.Context, id int64) error {
	if err := r.data.db.WithContext(ctx).Delete(&Artwork{}, id).Error; err != nil {
		r.log.Errorf("删除画作失败: %v", err)
		return err
	}
	r.log.Infof("画作删除成功: ID=%d", id)
	return nil
}

// FindByID 根据ID查找画作
func (r *artworkRepo) FindByID(ctx context.Context, id int64) (*biz.Artwork, error) {
	artworkModel := &Artwork{}
	if err := r.data.db.WithContext(ctx).
		Preload("Category").
		Preload("Tags").
		First(artworkModel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrArtworkNotFound
		}
		r.log.Errorf("查找画作失败: %v", err)
		return nil, err
	}
	return r.toBizArtwork(artworkModel), nil
}

// Search 搜索画作
func (r *artworkRepo) Search(ctx context.Context, req *biz.SearchRequest) ([]*biz.Artwork, int64, error) {
	query := r.data.db.WithContext(ctx).Model(&Artwork{}).
		Preload("Category").
		Preload("Tags")

	// 关键词搜索（标题、描述、作者）
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("title LIKE ? OR description LIKE ? OR author LIKE ?", keyword, keyword, keyword)
	}

	// 年份范围
	if req.YearFrom > 0 {
		query = query.Where("year >= ?", req.YearFrom)
	}
	if req.YearTo > 0 {
		query = query.Where("year <= ?", req.YearTo)
	}

	// 尺寸范围
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

	// 价格范围
	if req.PriceMin > 0 {
		query = query.Where("price >= ?", req.PriceMin)
	}
	if req.PriceMax > 0 {
		query = query.Where("price <= ?", req.PriceMax)
	}

	// 分类筛选
	if req.CategoryID > 0 {
		query = query.Where("category_id = ?", req.CategoryID)
	}

	// 标签筛选
	if len(req.Tags) > 0 {
		query = query.Joins("JOIN artwork_tags ON artworks.id = artwork_tags.artwork_id").
			Joins("JOIN tags ON artwork_tags.tag_id = tags.id").
			Where("tags.name IN ?", req.Tags).
			Group("artworks.id")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("统计画作数量失败: %v", err)
		return nil, 0, err
	}

	// 排序
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := strings.ToUpper(req.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// 分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	query = query.Offset(int(offset)).Limit(int(pageSize))

	// 查询结果
	var artworkModels []*Artwork
	if err := query.Find(&artworkModels).Error; err != nil {
		r.log.Errorf("搜索画作失败: %v", err)
		return nil, 0, err
	}

	artworks := make([]*biz.Artwork, 0, len(artworkModels))
	for _, model := range artworkModels {
		artworks = append(artworks, r.toBizArtwork(model))
	}

	return artworks, total, nil
}

// toBizArtwork 转换为业务模型
func (r *artworkRepo) toBizArtwork(model *Artwork) *biz.Artwork {
	artwork := &biz.Artwork{
		ID:          model.ID,
		Title:       model.Title,
		Description: model.Description,
		Author:      model.Author,
		Year:        model.Year,
		Width:       model.Width,
		Height:      model.Height,
		Material:    model.Material,
		Price:       model.Price,
		IsCollected: model.IsCollected,
		Location:    model.Location,
		Style:       model.Style,
		Technique:   model.Technique,
		ImageURL:    model.ImageURL,
		CategoryID:  model.CategoryID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	if model.Category != nil {
		artwork.Category = &biz.Category{
			ID:          model.Category.ID,
			Name:        model.Category.Name,
			Description: model.Category.Description,
		}
	}

	if len(model.Tags) > 0 {
		artwork.Tags = make([]*biz.Tag, 0, len(model.Tags))
		for _, tag := range model.Tags {
			artwork.Tags = append(artwork.Tags, &biz.Tag{
				ID:         tag.ID,
				Name:       tag.Name,
				UsageCount: tag.UsageCount,
			})
		}
	}

	return artwork
}
