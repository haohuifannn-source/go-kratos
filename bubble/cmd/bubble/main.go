package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"bubble/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	kratoszap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	f, _ := os.OpenFile("test.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	// zap日志库
	writeSyncer := zapcore.AddSync(f) // 写到哪里去

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()) // 用什么编码方法
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)   // 构建引擎
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
	// 直接用log方法，但是只有一个log方法
	//logger.Log(log.LevelDebug, "msg", "log init sucess")
	// 官方推荐用helper方法，因为具备很多的使用方式，且不用自己去传level
	helper := log.NewHelper(log.NewFilter(logger, log.FilterKey("password")))
	helper.Debug("log init sucess")
	helper.Infof("today is %v", time.Now())
	helper.Warnw("name", "lisa", "age", 18, "password", "12345")

	c := config.New(
		config.WithSource(
			env.NewSource("BUBBLE_"), // 指定环境变量前缀
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	fmt.Printf("---->bc.Server : %#v\n", bc.Server)
	fmt.Printf("---->bc.Data : %#v\n", bc.Data)
	fmt.Printf("---->bc.Mode : %#v\n", bc.Mode)

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
