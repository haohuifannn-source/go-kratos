package server

import (
	v1 "bubble/api/bubble/v1"
	"bubble/internal/conf"
	"bubble/internal/service"
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.TodoService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			selector.Server(
				Middleware(), // 加（）是为了可以根据传入的参数定制化
			).
				Path("/api.bubble.v1.Todo/CreateTodo").
				Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	//替换默认的HTTP响应
	opts = append(opts, http.ResponseEncoder(responseEncoder))
	opts = append(opts, http.ErrorEncoder(errorEncoder))

	srv := http.NewServer(opts...)
	v1.RegisterTodoHTTPServer(srv, greeter)
	return srv
}

// Middleware 自定义中间件
// type Middleware1 func(Handler) Handler
// type Handler func(ctx context.Context, req any) (any, error)
func Middleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			// 执行之前做点事
			fmt.Println("Middleware:执行handler之前")
			// 做token校验
			if tr, ok := transport.FromServerContext(ctx); ok {
				token := tr.RequestHeader().Get("token")
				fmt.Printf("token:%v\n", token)
			}
			defer func() {
				fmt.Println("Middleware: 执行handle之后")
			}()
			return handler(ctx, req)
		}
	}
}

// Middleware1 自定义中间件, 相比于下面的多了一些灵活性，可以定制化
// type Middleware1 func(Handler) Handler
func Middleware1(opts ...string) middleware.Middleware {
	return func(middleware.Handler) middleware.Handler {
		return nil
	}
}

// Middleware2 自定义中间件, 相比于上面的失去一些灵活性
// type Middleware2 func(Handler) Handler
func Middleware2(middleware.Handler) middleware.Middleware {
	return nil
}
