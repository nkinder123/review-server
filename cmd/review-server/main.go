package main

import (
	"flag"
	"fmt"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/hashicorp/consul/api"
	"os"
	"review-server/internal/conf"
	"review-server/pkg"

	kratoszap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "review.service"
	// Version is the version of the compiled software.
	Version string = "v1.0"
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

// 服务注册
func ReviewRegister(conf *conf.Register) *consul.Registry {
	c := api.DefaultConfig()
	c.Address = conf.Address
	c.Scheme = conf.Scheme
	client, err := api.NewClient(c)
	if err != nil {
		panic("register review-service has error")
	}
	return consul.New(client)
}

func newApp(logger log.Logger, r *conf.Register, gs *grpc.Server, hs *http.Server) *kratos.App {
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
		kratos.Registrar(ReviewRegister(r)),
	)
}

func main() {
	z, err := pkg.InitLog()
	if err != nil {
		fmt.Printf("init log has error")
	}
	flag.Parse()
	logger := log.With(
		kratoszap.NewLogger(z),
		log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		fmt.Printf("error has c.load")
		panic(err)
	}

	//初始化snowflake and bc conf
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	snowflake := bc.Snowflake
	app, cleanup, err := wireApp(bc.Server, bc.Elasticsearch, bc.Register, bc.Data, logger)
	pkg.Init(snowflake.GetStartTime(), snowflake.GetMachineId())

	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
