package data

import (
	"bubble/internal/biz"
	"bubble/internal/conf"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewDB, NewRedis, NewData, NewTodoRepo)

// Data .
type Data struct {
	// TODO wrapped database client
	db    *gorm.DB
	redis *redis.Client
}

// NewData .
func NewData(db *gorm.DB, red *redis.Client, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	// 如果直接在这里使用gorm链接DB，就不符合控制反转的要求
	// db, err := gorm.Open(mysql.Open(c.Database.Dsn), &gorm.Config{})
	// 正确方法应该采用依赖注入通过参数传进来
	return &Data{
		db:    db,
		redis: red,
	}, cleanup, nil
}

func NewDB(c *conf.Data) (*gorm.DB, error) {
	// 根据配置文件中指定的driver来链接不同的数据库
	switch strings.ToLower(c.Database.Driver) {
	case "mysql":
		db, err := gorm.Open(mysql.Open(c.Database.Dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		err = db.AutoMigrate(&biz.Todo{}) // 为了改动和表对齐
		if err != nil {
			return nil, err
		}
		return db, nil
	case "sqlite":
		db, err := gorm.Open(sqlite.Open(c.Database.Dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		err = db.AutoMigrate(&biz.Todo{})
		if err != nil {
			return nil, err
		}
		return db, nil
	}
	return nil, errors.New("invalid, parms")
}

func NewRedis(c *conf.Data) (*redis.Client, error) {
	// 根据配置文件中指定的driver来链接不同的数据库
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: "",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}
