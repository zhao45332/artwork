package data

import (
	"context"

	"artwork/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type categoryRepo struct {
	data *Data
	log  *log.Helper
}

// NewCategoryRepo 创建分类仓储
func NewCategoryRepo(data *Data, logger log.Logger) biz.CategoryRepo {
	return &categoryRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// FindByID 根据ID查找分类
func (r *categoryRepo) FindByID(ctx context.Context, id int64) (*biz.Category, error) {
	categoryModel := &Category{}
	if err := r.data.db.WithContext(ctx).First(categoryModel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrCategoryNotFound
		}
		r.log.Errorf("查找分类失败: %v", err)
		return nil, err
	}

	return &biz.Category{
		ID:          categoryModel.ID,
		Name:        categoryModel.Name,
		Description: categoryModel.Description,
	}, nil
}
