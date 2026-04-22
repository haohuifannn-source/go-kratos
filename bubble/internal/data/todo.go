package data

import (
	"context"
	"fmt"

	"bubble/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

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
	fmt.Printf("save: t : %#v\n", t)
	return t, nil
}

func (r *TodoRepo) Update(ctx context.Context, t *biz.Todo) error {
	return nil
}

func (r *TodoRepo) FindByID(context.Context, int64) (*biz.Todo, error) {
	return nil, nil
}

func (r *TodoRepo) Delete(context.Context, int64) error {
	return nil
}

func (r *TodoRepo) ListAll(context.Context) ([]*biz.Todo, error) {
	return nil, nil
}
