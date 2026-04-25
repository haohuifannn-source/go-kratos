# 通过一个bubble项目去熟悉go-kratos的框架

***注意***: 该框架采用的是DDD的逻辑，解耦了逻辑和数据库的操作

1. 首先通过
```bash
kratos new bubble
```
来得到一个初始化的项目

2. 然后通过
```bash
kratos proto add api/bubble/v1/todo.proto
```
来添加该项目所需要的proto文件。然后再去修改该文件。
```proto
import "google/api/annotations.proto";
```
一定不要忘了导入这个google文件

3. 通过命令在api下生成protoc代码
```bash
kratos proto client api/bubble/v1/todo.proto

make api ----在根目录下
```

4. 通过命令生成service下的代码
```bash
kratos proto server api/bubble/v1/todo.proto -t internal/service
```

5. 参考模板手动实现自己的biz层代码和data层代码
***重点***:这部分就是DDD的精髓，通过interface的方式抽象出必须实现的方法

6. 修改server层的代码，从而将api和service联系起来

7. 修改wire层代码，因为这部分代码不可以手动修改，因此通过命令去修改
```bash
cd cmd/bubble

wire
```

整个流程就是：server层会向api去注册我们的服务，而server层会带有service层的实体，而service会带有biz的实体，而biz可以去操作数据库。

***重点***：这里面设计了一个控制反转的思想，实现控制反转最常用的就是依赖注入。具体可以参考https://liwenzhou.com/posts/go/wire/#%E6%8E%A7%E5%88%B6%E5%8F%8D%E8%BD%AC%E4%B8%8E%E4%BE%9D%E8%B5%96%E6%B3%A8%E5%85%A5。为了不需要程序员自己去找和实现，利用了wire工具进行生成。

## 一个核心流程，例如添加一个模式或者数据库

1. 修改conf.proto文件，用pb定义配置
```proto
message Bootstrap {
  Server server = 1;
  Data data = 2;
  string Mode = 3; // 新加入的
}
```

2. 生成配置对应的Go代码
然后调用
```bash
make config
```
去更新conf.pb.go代码

3. 修改configs配置文件
在.yaml文件下添加对应的代码
```yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 1s
data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/test?parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
    read_timeout: 0.2s
    write_timeout: 0.2s
mode: dev //加入新加的
```

也可以用环境变量的配置方法，例如：addr: 0.0.0.0:${HTTP_PORT:8000}。
通过以下命令去设置环境变量
```bash
export BUBBLE_HTTP_PORT=8099
echo $BUBBLE_HTTP_PORT ---查看环境变量是否配置成功
```
然后在main函数里面
```bash
c := config.New(
		config.WithSource(
			env.NewSource("BUBBLE_"), // 指定环境变量前缀 ---一定要加前缀
			file.NewSource(flagconf),
		),
	)
```


***注意事项***:
1. pb文件要语法写对了；
2. 改完pb文件后一定要生成Go代码；
3. 配置文件要个Go代码里的结构体对应上；
4. 配置文件的语法要写对。

## 业务逻辑开发

开发思路:顺着请求的流程去写代码，service ---> biz ---> data
1. service: 服务的入口，实现API层定义的服务
2. biz: 业务逻辑层，复杂的业务逻辑都写在这里
3. dara: 数据层，数据有关的操作都在这里
4. 这里采用了Automigrate的方法去自动创建数据库。有则自动对齐，没有则创建

### 开发了增的接口

### 开发了更新的接口

1. 在go框架里面，更多还是采用.Model的做法，而不是泛接口的方式；
2. 在更新里面，updates会自动忽略零值，因此需要用map的方法；
3. 在调用里面要判断AffectRow是否起到了更新的作用

### 开发了删的接口

### 开发了获取列表的接口
1. 这里的重点是参数的返回和处理

***重点***: Post和PUT的区别在于幂等性：POST提交十次就会创建十次，PUT只会有创建一次，其他都是在原基础上修改。

## 自定义HTTP响应返回
可以覆盖默认的DefaultResponseEncoder，通过http.ResponseEncoder()配置，注入到http.Server()中

例子：
什么时候需要：
1. 需要对外提供一套HTTP接口；
2. 对外提供的RESTful接口要求有一套固定格式的响应数据
```json
{
  "code" : 200,
  "msg": "success",
  "data": []
}
```

### 自定义返回响应实现
1. 在server层替换HTTP响应的编码器
```go
// 自定义响应方法
func responseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return nil
	}
	// 判断是不是重定向
	if rd, ok := v.(kratoshttp.Redirector); ok {
		url, code := rd.Redirect()
		http.Redirect(w, r, url, code)
		return nil
	}
	// 构造自定义的响应结构体
	resp := &httpResponse{
		Code: http.StatusOK,
		Msg:  "sucess",
		Data: v,
	}

	codec, _ := kratoshttp.CodecForRequest(r, "Accept")
	data, err := codec.Marshal(resp) //json.Marshal
	if err != nil {
		return err
	}

	// 设置响应头：Content-Type： application/json
	w.Header().Set("Content-Type", "application/"+codec.Name())
	_, err = w.Write(data)
	return err
}
```

### 自定义错误响应
原理和自定义HTTP响应一样
```go
// errorEncoder 自定义错误响应编码器
func errorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	// 能从err里面解析出错误吗
	resp := new(httpResponse)
	// 能从err里面解析出错误码
	if gs, ok := status.FromError(err); ok {
		resp = &httpResponse{
			Code: kratostatus.FromGRPCCode(gs.Code()),
			Msg:  gs.Message(),
			Data: nil,
		}
	} else {
		resp = &httpResponse{
			Code: http.StatusInternalServerError, //500
			Msg:  "内部错误",
		}
	}
	codec, _ := kratoshttp.CodecForRequest(r, "Accept")
	body, err := codec.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/"+codec.Name())
	w.WriteHeader(resp.Code)
	_, _ = w.Write(body)
}
```

### 注册自定义编码器
在http.go中替换默认的编码器
```go
//替换默认的HTTP响应
	opts = append(opts, http.ResponseEncoder(responseEncoder))
	opts = append(opts, http.ErrorEncoder(errorEncoder))
```

### 自定义错误码枚举
1. 安装错误工具
```bash
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
```

2. 在api层定义自己的错误proto文件
```go
syntax = "proto3";

// 定义包名
package api.kratos.v1;
import "errors/errors.proto";

// 多语言特定包名，用于源代码引用
option go_package = "bubble/api/bubble/v1;v1";
option java_multiple_files = true;
option java_package = "api.bubble.v1";

enum ErrorReason {
  // 设置缺省错误码
  option (errors.default_code) = 500;

  // 为某个枚举单独设置错误码
  TODO_NOT_FOUND = 0 [(errors.code) = 404];

  INVALID_PARAM = 1 [(errors.code) = 400];
}
```

3. 添加相应的文件到makefile中
```MAKEFILE
protoc --proto_path=. \
         --proto_path=./third_party \
         --go_out=paths=source_relative:. \
         --go-errors_out=paths=source_relative:. \
         $(API_PROTO_FILES)

make errors ---执行
```

4. 然后在service调用服务的地方插入自定义的状态码
例如：在service的Get函数的地方返回状态码

## 日志的使用

### log日志的使用
1. 首先通过打开一个文件，并初始化
```go
f, _ := os.Open("test.log")
// 这是是指定打印的时候需要携带什么信息
logger := log.With(log.NewStdLogger(f),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
```

### zap日志库的使用
```go
// zap日志库
	writeSyncer := zapcore.AddSync(f)

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	z := zap.New(core)

	// log.With(log.NewStdLogger(f)
	logger := log.With(kratoszap.NewLogger(z),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
```

2. 然后就可以直接使用他的方法
```go
// 直接用log方法，但是只有一个log方法
	//logger.Log(log.LevelDebug, "msg", "log init sucess")
	// 官方推荐用helper方法，因为具备很多的使用方式，且不用自己去传level
	helper := log.NewHelper(logger)
	helper.Debug("log init sucess")
	helper.Infof("today is %v", time.Now())
	helper.Warnw("name", "lisa", "age", 18)
```

***重点***：zap日志库更偏向于给机器读，因为输出的时候json的格式，开发一般更偏向于用zap日志库

## 编译
1. 在根目录下创建一个文件夹bin
2. 执行
```bash
go build -o ./bin ./...

./bin/bubble -conf ./configs
```

## 中间件的使用

1. 定义在server层的http和grpc里面
```go
func Middleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			// 执行之前做点事
			fmt.Println("Middleware:执行handler之前")
			defer func() {
				fmt.Println("Middleware: 执行handle之后")
			}()
			return handler(ctx, req)
		}
	}
}
```
2. 然后添加相应的代码
```go
var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			selector.Server( -----------通过selector可以为部分的http响应添加对应的中间件
				Middleware(), // 加（）是为了可以根据传入的参数定制化
			).
				Path("/hello.Update/UpdateUser").
				Build(),
		),
	}
```
***注意***：Path的路径的拼接规则是： /包名.服务名/方法名。例如下面的Create方法，就写/api.bubble.v1.Todo/CreateTodo，无论是grpc还是http

```proto
syntax = "proto3";

package api.bubble.v1;

import "google/api/annotations.proto";

option go_package = "bubble/api/bubble/v1;v1";
option java_multiple_files = true;
option java_package = "api.bubble.v1";

service Todo {
	rpc CreateTodo (CreateTodoRequest) returns (CreateTodoReply){
		option(google.api.http) = {
			post : "/v1/todo",
			body : "*"
		};
	};
	rpc UpdateTodo (UpdateTodoRequest) returns (UpdateTodoReply){
		option(google.api.http) = {
			put : "/v1/todo/{id}",
			body : "*"
		};
	};
	rpc DeleteTodo (DeleteTodoRequest) returns (DeleteTodoReply){
		option(google.api.http) = {
			delete : "/v1/todo/{id}",
		};
	};
	rpc GetTodo (GetTodoRequest) returns (GetTodoReply){
		option(google.api.http) = {
			get : "/v1/todo/{id}",
		};
	};
	rpc ListTodo (ListTodoRequest) returns (ListTodoReply){
		option(google.api.http) = {
			get : "/v1/todos",
		};
	};
}

message todo{
	int64 id = 1;
	string title = 2;
	bool status = 3;
}

message CreateTodoRequest {
	string title = 1;
}
message CreateTodoReply {
	todo todo = 1;
}

message UpdateTodoRequest {
	int64 id = 1;
	string title = 2;
	bool status = 3;
}
message UpdateTodoReply {
	string message = 1;
}

message DeleteTodoRequest {
	int64 id = 1;
}
message DeleteTodoReply {
	string message = 1;
}

message GetTodoRequest {
	int64 id =1;
}
message GetTodoReply {
	todo todo = 1;
}

message ListTodoRequest {

}
message ListTodoReply {
	repeated todo data = 1;
}
```

