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
