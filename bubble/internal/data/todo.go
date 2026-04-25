package data

import (
	"context"
	"errors"

	"bubble/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// TodoRepo实现了biz层定义的repo接口
type TodoRepo struct {
	data *Data
	log  *log.Helper
}

// NewTodoRepo .
func NewTodoRepo(data *Data, logger log.Logger) biz.TodoRepo {
	return &TodoRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *TodoRepo) Save(ctx context.Context, t *biz.Todo) (*biz.Todo, error) {
	// 数据库的操作
	err := r.data.db.Create(t).Error
	return t, err
}

func (r *TodoRepo) Update(ctx context.Context, t *biz.Todo) error {
	result := r.data.db.Model(&biz.Todo{}).Where("id = ?", t.ID).
		Updates(map[string]interface{}{
			"title":  t.Title,
			"status": t.Status,
		}) // 通过updatas可以自动忽略零值
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("没有这个todo")
	}
	return nil
}

func (r *TodoRepo) FindByID(ctx context.Context, id int64) (*biz.Todo, error) {
	t := &biz.Todo{ID: id}
	err := r.data.db.First(t).Error
	return t, err
}

func (r *TodoRepo) Delete(ctx context.Context, id int64) error {
	res := r.data.db.WithContext(ctx).Where("id=?", id).Delete(&biz.Todo{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("没有该id的todo")
	}
	return nil
}

func (r *TodoRepo) ListAll(context.Context) ([]*biz.Todo, error) {
	// 1. 定义与返回值类型一致的切片
	var res []*biz.Todo

	// 2. 传入指针 &res，并捕获 Error
	if err := r.data.db.Find(&res).Error; err != nil {
		// 如果查询出错，返回 nil 和错误信息
		return nil, err
	}

	// 3. 返回查询到的结果
	return res, nil
}
