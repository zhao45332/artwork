package service

import (
	"context"

	v1 "artwork/api/artwork/v1"
	"artwork/internal/biz"
	"artwork/internal/pkg/xctx"

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

func (s *ArtworkService) CreateArtwork(ctx context.Context, req *v1.CreateArtworkRequest) (*v1.CreateArtworkReply, error) {
	artwork := s.fromCreateRequest(req)
	result, err := s.uc.CreateArtwork(ctx, artwork)
	if err != nil {
		return nil, err
	}
	return &v1.CreateArtworkReply{Artwork: s.toArtworkInfo(result)}, nil
}

func (s *ArtworkService) UpdateArtwork(ctx context.Context, req *v1.UpdateArtworkRequest) (*v1.UpdateArtworkReply, error) {
	artwork := s.fromUpdateRequest(req)
	result, err := s.uc.UpdateArtwork(ctx, artwork)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateArtworkReply{Artwork: s.toArtworkInfo(result)}, nil
}

func (s *ArtworkService) DeleteArtwork(ctx context.Context, req *v1.DeleteArtworkRequest) (*v1.DeleteArtworkReply, error) {
	if err := s.uc.DeleteArtwork(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteArtworkReply{Success: true}, nil
}

func (s *ArtworkService) GetArtwork(ctx context.Context, req *v1.GetArtworkRequest) (*v1.GetArtworkReply, error) {
	artwork, err := s.uc.GetArtwork(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetArtworkReply{Artwork: s.toArtworkInfo(artwork)}, nil
}

func (s *ArtworkService) SearchArtwork(ctx context.Context, req *v1.SearchArtworkRequest) (*v1.SearchArtworkReply, error) {
	artworks, total, err := s.uc.SearchArtwork(ctx, &biz.SearchRequest{
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
		UserID:     req.UserId,
		Visibility: req.Visibility,
		Status:     req.Status,
	})
	if err != nil {
		return nil, err
	}
	return s.toSearchReply(artworks, total, req.Page, req.PageSize), nil
}

func (s *ArtworkService) GetMyArtworks(ctx context.Context, req *v1.GetMyArtworksRequest) (*v1.SearchArtworkReply, error) {
	artworks, total, err := s.uc.GetMyArtworks(ctx, req.Page, req.PageSize, req.Status)
	if err != nil {
		return nil, err
	}
	return s.toSearchReply(artworks, total, req.Page, req.PageSize), nil
}

func (s *ArtworkService) UploadImage(ctx context.Context, req *v1.UploadImageRequest) (*v1.UploadImageReply, error) {
	contentType := req.ContentType
	if contentType == "" {
		contentType = "image/jpeg"
	}
	url, err := s.uc.UploadImage(ctx, req.Filename, req.Content, contentType)
	if err != nil {
		return nil, err
	}
	return &v1.UploadImageReply{Url: url, Filename: req.Filename}, nil
}

func (s *ArtworkService) fromCreateRequest(req *v1.CreateArtworkRequest) *biz.Artwork {
	artwork := &biz.Artwork{
		Title:         req.Title,
		Description:   req.Description,
		Author:        req.Author,
		Year:          req.Year,
		Width:         req.Width,
		Height:        req.Height,
		Material:      req.Material,
		Price:         req.Price,
		IsCollected:   req.IsCollected,
		Location:      req.Location,
		Style:         req.Style,
		Technique:     req.Technique,
		ImageURL:      req.ImageUrl,
		CoverImageURL: req.CoverImageUrl,
		ImageURLs:     req.ImageUrls,
		CategoryID:    req.CategoryId,
		Status:        req.Status,
		Visibility:    req.Visibility,
	}
	if len(req.Tags) > 0 {
		artwork.Tags = make([]*biz.Tag, 0, len(req.Tags))
		for _, tagName := range req.Tags {
			artwork.Tags = append(artwork.Tags, &biz.Tag{Name: tagName})
		}
	}
	return artwork
}

func (s *ArtworkService) fromUpdateRequest(req *v1.UpdateArtworkRequest) *biz.Artwork {
	artwork := s.fromCreateRequest(&v1.CreateArtworkRequest{
		Title:         req.Title,
		Description:   req.Description,
		Author:        req.Author,
		Year:          req.Year,
		Width:         req.Width,
		Height:        req.Height,
		Material:      req.Material,
		Price:         req.Price,
		IsCollected:   req.IsCollected,
		Location:      req.Location,
		Style:         req.Style,
		Technique:     req.Technique,
		ImageUrl:      req.ImageUrl,
		CoverImageUrl: req.CoverImageUrl,
		ImageUrls:     req.ImageUrls,
		CategoryId:    req.CategoryId,
		Tags:          req.Tags,
		Status:        req.Status,
		Visibility:    req.Visibility,
	})
	artwork.ID = req.Id
	return artwork
}

func (s *ArtworkService) toSearchReply(artworks []*biz.Artwork, total int64, page, pageSize int32) *v1.SearchArtworkReply {
	artworkInfos := make([]*v1.ArtworkInfo, 0, len(artworks))
	for _, artwork := range artworks {
		artworkInfos = append(artworkInfos, s.toArtworkInfo(artwork))
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 12
	}
	return &v1.SearchArtworkReply{Artworks: artworkInfos, Total: int32(total), Page: page, PageSize: pageSize}
}

func (s *ArtworkService) toArtworkInfo(artwork *biz.Artwork) *v1.ArtworkInfo {
	info := &v1.ArtworkInfo{
		Id:            artwork.ID,
		UserId:        artwork.UserID,
		Username:      artwork.Username,
		UserNickname:  artwork.UserNickname,
		UserAvatarUrl: artwork.UserAvatarURL,
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
		ImageUrl:      artwork.ImageURL,
		CoverImageUrl: artwork.CoverImageURL,
		CategoryId:    artwork.CategoryID,
		Status:        artwork.Status,
		Visibility:    artwork.Visibility,
		ViewCount:     artwork.ViewCount,
		LikeCount:     artwork.LikeCount,
		FavoriteCount: artwork.FavoriteCount,
		CommentCount:  artwork.CommentCount,
		ShareCount:    artwork.ShareCount,
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
	if len(artwork.ImageURLs) > 0 {
		info.Images = make([]*v1.ArtworkImage, 0, len(artwork.ImageURLs))
		for idx, url := range artwork.ImageURLs {
			info.Images = append(info.Images, &v1.ArtworkImage{Id: int64(idx + 1), Url: url, IsCover: url == artwork.CoverImageURL, SortOrder: int32(idx)})
		}
	}
	if artwork.PublishAt != nil {
		info.PublishAt = timestamppb.New(*artwork.PublishAt)
	}
	if !artwork.CreatedAt.IsZero() {
		info.CreatedAt = timestamppb.New(artwork.CreatedAt)
	}
	if !artwork.UpdatedAt.IsZero() {
		info.UpdatedAt = timestamppb.New(artwork.UpdatedAt)
	}
	return info
}

func currentUserID(ctx context.Context) int64 {
	userID, _ := xctx.UserIDFromContext(ctx)
	return userID
}
