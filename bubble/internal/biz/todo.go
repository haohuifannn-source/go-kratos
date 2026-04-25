package biz

import (
	"context"
	"fmt"
	"time"

	v1 "bubble/api/helloworld/v1"
	"bubble/internal/conf"
	"bubble/third_party/auth"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	// ErrTodoNotFound is user not found.
	ErrTodoNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// Todo is a todo model.
type Todo struct {
	ID     int64
	Title  string
	Status bool
}

// TodoRepo is a Todo repo.
// biz层对数据操作层提出了以下的要求，不关心实际的存储是mysql还是redis还是MongoDB
type TodoRepo interface {
	Save(context.Context, *Todo) (*Todo, error)
	Update(context.Context, *Todo) error
	Delete(context.Context, int64) error
	FindByID(context.Context, int64) (*Todo, error)
	ListAll(context.Context) ([]*Todo, error)

	//redis的操作
	SetRefreshToken(context.Context, int64, string, time.Duration) error
}

// TodoUsecase is a todo usecase.
type TodoUsecase struct {
	repo TodoRepo
	log  *log.Helper
	conf *conf.Token
}

// NewTodoUsecase new a todo usecase.
func NewTodoUsecase(repo TodoRepo, logger log.Logger, c *conf.Token) *TodoUsecase {
	return &TodoUsecase{repo: repo, log: log.NewHelper(logger), conf: c}
}

// CreateTodo creates a Todo, and returns the new Todo.
// 对外提供的业务函数，实现复杂的业务逻辑
func (uc *TodoUsecase) CreateTodo(ctx context.Context, t *Todo) (*Todo, error) {
	uc.log.WithContext(ctx).Infof("Create: %#v", t)
	// 1. 生成token
	accessScre := uc.conf.AccessToken.AccessSecret
	accessExp := uc.conf.AccessToken.AccessExpire
	accessToken, err := auth.GenerateToken(accessScre, time.Now().Unix(), accessExp, t.ID)
	if err != nil {
		uc.log.Errorf("generate accessToken failed : %v\n", err)
		return nil, err
	}
	refreshScre := uc.conf.RefreshToken.RefreshSecret
	refreshExp := uc.conf.RefreshToken.RefreshExpire
	refreshToken, err := auth.GenerateToken(refreshScre, time.Now().Unix(), refreshExp, t.ID)
	if err != nil {
		uc.log.Errorf("generate refreshToken failed : %v\n", err)
		return nil, err
	}
	// 2. 将生成的AccessToken写进上下文里面
	fmt.Printf("----->AccessToken is %v\n", accessToken)
	fmt.Printf("----->refreshToken is %v\n", refreshToken)
	fmt.Printf("----->t.ID is %v\n", t.ID)
	// 3. 存储数据，并将生成的RefreshToken放进redis中
	u, err := uc.repo.Save(ctx, t)
	if err != nil {
		// 如果 Redis 存入失败，直接拦截，防止后续逻辑产生孤儿数据
		return nil, errors.InternalServer("uc.repo.Save", "存储失败")
	}
	err = uc.repo.SetRefreshToken(ctx, u.ID, refreshToken, time.Duration(refreshExp)*time.Second)
	if err != nil {
		// 如果 Redis 存入失败，直接拦截，防止后续逻辑产生孤儿数据
		return nil, errors.InternalServer("REDIS_ERROR", "存储失败")
	}
	return u, err // 调用save函数
}

// GetTodo creates a Todo, and returns the new Todo.
// 对外提供的业务函数，实现复杂的业务逻辑
func (uc *TodoUsecase) GetTodo(ctx context.Context, id int64) (*Todo, error) {
	uc.log.WithContext(ctx).Infof("Get: %#v", id)
	return uc.repo.FindByID(ctx, id) // 调用save函数
}

// UpdataTodo creates a Todo, and returns the new Todo.
// 对外提供的业务函数，实现复杂的业务逻辑
func (uc *TodoUsecase) UpdataTodo(ctx context.Context, t *Todo) error {
	uc.log.WithContext(ctx).Infof("Put: %#v", t)
	return uc.repo.Update(ctx, t) // 调用save函数
}

// DeleteTodo creates a Todo, and returns the new Todo.
// 对外提供的业务函数，实现复杂的业务逻辑
func (uc *TodoUsecase) DeleteTodo(ctx context.Context, id int64) error {
	uc.log.WithContext(ctx).Infof("Delete: %#v", id)
	return uc.repo.Delete(ctx, id) // 调用save函数
}

// DeleteTodo creates a Todo, and returns the new Todo.
// 对外提供的业务函数，实现复杂的业务逻辑
func (uc *TodoUsecase) ListTodo(ctx context.Context) ([]*Todo, error) {
	uc.log.WithContext(ctx).Infof("Get: %#v")
	return uc.repo.ListAll(ctx) // 调用save函数
}
