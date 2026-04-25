package service

import (
	"context"
	"errors"

	pb "bubble/api/bubble/v1"
	"bubble/internal/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TodoService struct {
	pb.UnimplementedTodoServer

	// 嵌入一个实现业务逻辑的结构体
	uc *biz.TodoUsecase
}

func NewTodoService(uc *biz.TodoUsecase) *TodoService {
	return &TodoService{
		uc: uc,
	}
}

func (s *TodoService) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.CreateTodoReply, error) {
	// 请求来了，做参数的校验....
	if len(req.Title) == 0 {
		return &pb.CreateTodoReply{}, errors.New("无效的title")
	}
	// 调用逻辑
	data, err := s.uc.CreateTodo(ctx, &biz.Todo{
		Title: req.Title,
	})
	if err != nil {
		return nil, errors.New("内部错误")
	}
	// 返回调用的结果
	return &pb.CreateTodoReply{
		Todo: &pb.Todo{
			Id:     data.ID,
			Status: data.Status,
			Title:  data.Title,
		},
	}, nil
}
func (s *TodoService) UpdateTodo(ctx context.Context, req *pb.UpdateTodoRequest) (*pb.UpdateTodoReply, error) {
	// 参数校验
	if req.Id <= 0 {
		return nil, errors.New("参数错误")
	}
	// 调用逻辑处理
	todo := &biz.Todo{
		ID:     req.Id,
		Title:  req.Title,
		Status: req.Status,
	}
	if err := s.uc.UpdataTodo(ctx, todo); err != nil {
		return nil, errors.New("内部错误")
	}
	// 返回结果
	return &pb.UpdateTodoReply{
		Message: "更新成功！",
	}, nil
}

func (s *TodoService) DeleteTodo(ctx context.Context, req *pb.DeleteTodoRequest) (*pb.DeleteTodoReply, error) {
	if req.Id <= 0 {
		return nil, errors.New("参数错误")
	}
	err := s.uc.DeleteTodo(ctx, req.Id)
	if err != nil {
		return nil, errors.New("内部错误")
	}
	return &pb.DeleteTodoReply{
		Message: "删除成功",
	}, nil
}
func (s *TodoService) GetTodo(ctx context.Context, req *pb.GetTodoRequest) (*pb.GetTodoReply, error) {
	// 参数校验
	if req.Id <= 0 {
		return nil, errors.New("参数错误")
	}
	// 调用查询的逻辑
	t, err := s.uc.GetTodo(ctx, req.Id)
	if err != nil {
		//return nil, errors.New("内部错误")
		// 返回自定义错误码
		return nil, pb.ErrorTodoNotFound("id:%v todo is not found", req.Id)
	}
	// 返回参数
	return &pb.GetTodoReply{
		Todo: &pb.Todo{
			Id:     t.ID,
			Title:  t.Title,
			Status: t.Status,
		},
	}, nil
}
func (s *TodoService) ListTodo(ctx context.Context, req *pb.ListTodoRequest) (*pb.ListTodoReply, error) {
	res, err := s.uc.ListTodo(ctx)
	if err != nil {
		return nil, errors.New("内部错误")
	}
	data := make([]*pb.Todo, 0, len(res))
	for _, v := range res {
		data = append(data, &pb.Todo{
			Id:     v.ID,
			Title:  v.Title,
			Status: v.Status,
		})
	}
	return &pb.ListTodoReply{
		Data: data,
	}, nil
}

func (s *TodoService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenReply, error) {

	return nil, status.Error(codes.Unimplemented, "method RefreshToken not implemented")
}
