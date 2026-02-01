package service

import (
	"context"

	v1 "artwork/api/artwork/v1"
	"artwork/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ArtworkService 画作服务
type ArtworkService struct {
	v1.UnimplementedArtworkServiceServer

	uc  *biz.ArtworkUsecase
	log *log.Helper
}

// NewArtworkService 创建画作服务
func NewArtworkService(uc *biz.ArtworkUsecase, logger log.Logger) *ArtworkService {
	return &ArtworkService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// CreateArtwork 创建画作
func (s *ArtworkService) CreateArtwork(ctx context.Context, req *v1.CreateArtworkRequest) (*v1.CreateArtworkReply, error) {
	artwork := &biz.Artwork{
		Title:       req.Title,
		Description: req.Description,
		Author:      req.Author,
		Year:        req.Year,
		Width:       req.Width,
		Height:      req.Height,
		Material:    req.Material,
		Price:       req.Price,
		IsCollected: req.IsCollected,
		Location:    req.Location,
		Style:       req.Style,
		Technique:   req.Technique,
		ImageURL:    req.ImageUrl,
		CategoryID:  req.CategoryId,
	}

	// 处理标签
	if len(req.Tags) > 0 {
		artwork.Tags = make([]*biz.Tag, 0, len(req.Tags))
		for _, tagName := range req.Tags {
			artwork.Tags = append(artwork.Tags, &biz.Tag{Name: tagName})
		}
	}

	result, err := s.uc.CreateArtwork(ctx, artwork)
	if err != nil {
		return nil, err
	}

	return &v1.CreateArtworkReply{
		Artwork: s.toArtworkInfo(result),
	}, nil
}

// UpdateArtwork 更新画作
func (s *ArtworkService) UpdateArtwork(ctx context.Context, req *v1.UpdateArtworkRequest) (*v1.UpdateArtworkReply, error) {
	artwork := &biz.Artwork{
		ID:          req.Id,
		Title:       req.Title,
		Description: req.Description,
		Author:      req.Author,
		Year:        req.Year,
		Width:       req.Width,
		Height:      req.Height,
		Material:    req.Material,
		Price:       req.Price,
		IsCollected: req.IsCollected,
		Location:    req.Location,
		Style:       req.Style,
		Technique:   req.Technique,
		ImageURL:    req.ImageUrl,
		CategoryID:  req.CategoryId,
	}

	// 处理标签
	if len(req.Tags) > 0 {
		artwork.Tags = make([]*biz.Tag, 0, len(req.Tags))
		for _, tagName := range req.Tags {
			artwork.Tags = append(artwork.Tags, &biz.Tag{Name: tagName})
		}
	}

	result, err := s.uc.UpdateArtwork(ctx, artwork)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateArtworkReply{
		Artwork: s.toArtworkInfo(result),
	}, nil
}

// DeleteArtwork 删除画作
func (s *ArtworkService) DeleteArtwork(ctx context.Context, req *v1.DeleteArtworkRequest) (*v1.DeleteArtworkReply, error) {
	err := s.uc.DeleteArtwork(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteArtworkReply{
		Success: true,
	}, nil
}

// GetArtwork 获取画作
func (s *ArtworkService) GetArtwork(ctx context.Context, req *v1.GetArtworkRequest) (*v1.GetArtworkReply, error) {
	artwork, err := s.uc.GetArtwork(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.GetArtworkReply{
		Artwork: s.toArtworkInfo(artwork),
	}, nil
}

// SearchArtwork 搜索画作
func (s *ArtworkService) SearchArtwork(ctx context.Context, req *v1.SearchArtworkRequest) (*v1.SearchArtworkReply, error) {
	searchReq := &biz.SearchRequest{
		Keyword:    req.Keyword,
		YearFrom:   req.YearFrom,
		YearTo:     req.YearTo,
		WidthMin:   req.WidthMin,
		WidthMax:   req.WidthMax,
		HeightMin:  req.HeightMin,
		HeightMax:  req.HeightMax,
		PriceMin:   req.PriceMin,
		PriceMax:   req.PriceMax,
		CategoryID: req.CategoryId,
		Tags:       req.Tags,
		Page:       req.Page,
		PageSize:   req.PageSize,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
	}

	artworks, total, err := s.uc.SearchArtwork(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	artworkInfos := make([]*v1.ArtworkInfo, 0, len(artworks))
	for _, artwork := range artworks {
		artworkInfos = append(artworkInfos, s.toArtworkInfo(artwork))
	}

	return &v1.SearchArtworkReply{
		Artworks: artworkInfos,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UploadImage 上传图片
func (s *ArtworkService) UploadImage(ctx context.Context, req *v1.UploadImageRequest) (*v1.UploadImageReply, error) {
	s.log.Infof("正在上传图片: %s, 大小: %d 字节", req.Filename, len(req.Content))

	contentType := req.ContentType
	if contentType == "" {
		contentType = "image/jpeg" // 默认类型
	}

	url, err := s.uc.UploadImage(ctx, req.Filename, req.Content, contentType)
	if err != nil {
		s.log.Errorf("上传图片失败: %v", err)
		return nil, err
	}

	return &v1.UploadImageReply{
		Url:      url,
		Filename: req.Filename,
	}, nil
}

// toArtworkInfo 转换为 API 模型
func (s *ArtworkService) toArtworkInfo(artwork *biz.Artwork) *v1.ArtworkInfo {
	info := &v1.ArtworkInfo{
		Id:          artwork.ID,
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
		ImageUrl:    artwork.ImageURL,
		CategoryId:  artwork.CategoryID,
	}

	if artwork.Category != nil {
		info.CategoryName = artwork.Category.Name
	}

	if len(artwork.Tags) > 0 {
		info.Tags = make([]string, 0, len(artwork.Tags))
		for _, tag := range artwork.Tags {
			info.Tags = append(info.Tags, tag.Name)
		}
	}

	// 设置时间戳
	if !artwork.CreatedAt.IsZero() {
		info.CreatedAt = timestamppb.New(artwork.CreatedAt)
	}
	if !artwork.UpdatedAt.IsZero() {
		info.UpdatedAt = timestamppb.New(artwork.UpdatedAt)
	}

	return info
}
