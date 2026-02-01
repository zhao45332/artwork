package data

import (
	"context"

	"artwork/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type tagRepo struct {
	data *Data
	log  *log.Helper
}

// NewTagRepo 创建标签仓储
func NewTagRepo(data *Data, logger log.Logger) biz.TagRepo {
	return &tagRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// FindOrCreateByName 根据名称查找或创建标签
func (r *tagRepo) FindOrCreateByName(ctx context.Context, name string) (*biz.Tag, error) {
	tagModel := &Tag{}
	err := r.data.db.WithContext(ctx).Where("name = ?", name).First(tagModel).Error
	if err == nil {
		return &biz.Tag{
			ID:         tagModel.ID,
			Name:       tagModel.Name,
			UsageCount: tagModel.UsageCount,
		}, nil
	}

	if err != gorm.ErrRecordNotFound {
		r.log.Errorf("查找标签失败: %v", err)
		return nil, err
	}

	// 创建新标签
	tagModel = &Tag{
		Name:       name,
		UsageCount: 0,
	}
	if err := r.data.db.WithContext(ctx).Create(tagModel).Error; err != nil {
		r.log.Errorf("创建标签失败: %v", err)
		return nil, err
	}

	return &biz.Tag{
		ID:         tagModel.ID,
		Name:       tagModel.Name,
		UsageCount: tagModel.UsageCount,
	}, nil
}

// FindByNames 根据名称列表查找标签
func (r *tagRepo) FindByNames(ctx context.Context, names []string) ([]*biz.Tag, error) {
	if len(names) == 0 {
		return nil, nil
	}

	var tagModels []*Tag
	if err := r.data.db.WithContext(ctx).Where("name IN ?", names).Find(&tagModels).Error; err != nil {
		r.log.Errorf("查找标签失败: %v", err)
		return nil, err
	}

	// 创建不存在的标签
	existingNames := make(map[string]bool)
	tags := make([]*biz.Tag, 0, len(names))
	for _, model := range tagModels {
		existingNames[model.Name] = true
		tags = append(tags, &biz.Tag{
			ID:         model.ID,
			Name:       model.Name,
			UsageCount: model.UsageCount,
		})
	}

	// 创建不存在的标签
	for _, name := range names {
		if !existingNames[name] {
			tag, err := r.FindOrCreateByName(ctx, name)
			if err != nil {
				return nil, err
			}
			tags = append(tags, tag)
		}
	}

	return tags, nil
}
