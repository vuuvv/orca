package orca

import (
	"context"
	"database/sql"
	redis2 "github.com/go-redis/redis/v8"
	"github.com/meehow/securebytes"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/orca/config"
	"github.com/vuuvv/orca/id"
	"github.com/vuuvv/orca/logger"
	"github.com/vuuvv/orca/orm"
	"github.com/vuuvv/orca/redis"
	"github.com/vuuvv/orca/redislock"
	"github.com/vuuvv/orca/secure"
	"github.com/vuuvv/orca/serialize"
	"github.com/vuuvv/orca/server"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type Application struct {
	configPath   string
	configLoader config.Loader
	httpServer   server.Server
	logger       *zap.Logger
	redisClient  *redis2.Client
	db           *gorm.DB
	idGenerator  *id.Generator
	secure       *securebytes.SecureBytes
}

type ApplicationOption func(app *Application)

func WithConfigLoader(loader config.Loader) ApplicationOption {
	return func(app *Application) {
		app.configLoader = loader
	}
}

func WithConfigPath(path string) ApplicationOption {
	return func(app *Application) {
		app.configPath = path
	}
}

func WithHttpServer(server *server.GinServer) ApplicationOption {
	return func(app *Application) {
		app.httpServer = server
	}
}

func WithLogger(logger *zap.Logger) ApplicationOption {
	return func(app *Application) {
		app.logger = logger
	}
}

func NewApplication(opts ...ApplicationOption) *Application {
	app := &Application{
		configPath:   "resources/application.yaml",
		configLoader: config.NewViperConfigLoader(),
	}
	for _, opt := range opts {
		opt(app)
	}

	// jsoniter序列化初始化
	serialize.InitializeJsoniter()

	// 配置加载器初始化
	err := app.configLoader.Load(app.configPath)
	if err != nil {
		panic(err)
	}

	// 日志初始化
	zapConfig := &logger.Config{}
	err = app.UnmarshalConfig(zapConfig, "zap")
	if err != nil {
		panic(err)
	}
	app.logger = logger.NewLogger(zapConfig)

	// redis初始化
	if app.configLoader.IsSet("redis") {
		redisConfig := &redis.Config{}
		err = app.UnmarshalConfig(redisConfig, "redis")
		if err != nil {
			panic(err)
		}
		app.redisClient, err = redis.NewRedisClient(redisConfig)
		if err != nil {
			panic(err)
		}
		redis.SetClient(app.redisClient)
	}

	if app.redisClient != nil {
		app.idGenerator, err = id.NewGenerator(id.WithRedisClient(app.redisClient))
		if err != nil {
			panic(err)
		}
	} else {
		app.idGenerator, err = id.NewGenerator()
	}

	// gorm初始化
	if app.configLoader.IsSet("database") {
		databaseConfig := &orm.Config{}
		err = app.UnmarshalConfig(databaseConfig, "database")
		if err != nil {
			panic(err)
		}
		app.db, err = orm.New(databaseConfig)
		if err != nil {
			panic(err)
		}
	}

	// http服务器初始化
	httpConfig := &server.Config{}
	err = app.UnmarshalConfig(httpConfig, "http")
	if err != nil {
		panic(err)
	}
	app.httpServer = server.NewGinServer(httpConfig)

	if defaultApplication == nil {
		ReplaceDefaultApplication(app)
	}

	secure.SetSecure(secure.NewSecure(httpConfig.JwtSecret))

	return app
}

func (this *Application) GetConfig(name string) interface{} {
	return this.configLoader.Get(name)
}

func (this *Application) UnmarshalConfig(output interface{}, name ...string) error {
	return this.configLoader.Unmarshal(output, name...)
}

func (this *Application) Mount(controllers ...interface{}) *Application {
	this.httpServer.Mount(controllers...)
	return this
}

func (this *Application) SetDefault() {
	ReplaceDefaultApplication(this)
}

func (this *Application) GetHttpServer() server.Server {
	return this.httpServer
}

func (this *Application) Start() {
	if this.httpServer == nil {
		panic("Application start error: http server is nil")
	}

	this.httpServer.Start()
}

func (this *Application) Use(handlers ...interface{}) *Application {
	if this.httpServer == nil {
		panic("Application start error: http server is nil")
	}
	this.httpServer.Use(handlers...)
	return this
}

func (this *Application) Default() server.Server {
	if this.httpServer == nil {
		panic("Application start error: http server is nil")
	}
	return this.httpServer.Default()
}

var defaultApplication *Application = nil

func App() *Application {
	return defaultApplication
}

func ReplaceDefaultApplication(app *Application) {
	defaultApplication = app
}

func GetConfig(name string) interface{} {
	return defaultApplication.GetConfig(name)
}

func UnmarshalConfig(output interface{}, name ...string) error {
	return defaultApplication.UnmarshalConfig(output, name...)
}

func Redis() *redis2.Client {
	return defaultApplication.redisClient
}

func Database() *gorm.DB {
	return defaultApplication.db
}

func RedisLock(ctx context.Context, key string, ttl time.Duration, opt *redislock.Options) (*redislock.Lock, error) {
	locker := redislock.New(Redis())
	lock, err := locker.Obtain(ctx, key, ttl, opt)
	return lock, errors.WithStack(err)
}

func Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return errors.WithStack(defaultApplication.db.Transaction(fc, opts...))
}

func Use(handlers ...interface{}) *Application {
	if defaultApplication == nil {
		defaultApplication = NewApplication()
	}
	defaultApplication.Use(handlers...)
	return defaultApplication
}

func Start(controllers ...interface{}) {
	if defaultApplication == nil {
		defaultApplication = NewApplication()
	}
	defaultApplication.Mount(controllers...).Start()
}

func StartDefault(controllers ...interface{}) {
	if defaultApplication == nil {
		defaultApplication = NewApplication()
	}
	defaultApplication.Default().Mount(controllers...).Start()
}
