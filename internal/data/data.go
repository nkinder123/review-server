package data

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"review-server/internal/conf"
	"review-server/internal/data/query"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewReviewRepo, Connect, NewElasticSearch, NewRedsiClient)

// Data .
type Data struct {
	// TODO wrapped database client
	query         *query.Query
	elasticClient *elasticsearch.TypedClient
	log           *log.Helper
	rdb           *redis.Client
}

// NewData .
func NewData(client *elasticsearch.TypedClient, rdb *redis.Client, db *gorm.DB, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db)
	return &Data{
		query:         query.Q,
		elasticClient: client,
		log:           log.NewHelper(logger),
		rdb:           rdb,
	}, cleanup, nil
}

func NewElasticSearch(cfg *conf.Elasticsearch) (*elasticsearch.TypedClient, error) {
	c := elasticsearch.Config{
		Addresses: cfg.Addresses,
	}
	return elasticsearch.NewTypedClient(c)
}

func NewRedsiClient(cfg *conf.Data) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		ReadTimeout:  cfg.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: cfg.Redis.WriteTimeout.AsDuration(),
	})
	return client
}

func Connect(cfg *conf.Data) *gorm.DB {
	if cfg == nil {
		log.Fatal("config db is nil")
		return nil
	}
	switch strings.ToLower(cfg.Database.Driver) {
	case "mysql":
		db, err := gorm.Open(mysql.Open(cfg.Database.Source))
		if err != nil {
			log.Fatal("mysql db connect has error:", err)
			return nil
		}
		return db
	case "sqlite":
		db, err := gorm.Open(sqlite.Open(cfg.Database.Source))
		if err != nil {
			log.Fatal("sqlite connect has error")
			return nil
		}
		return db
	}
	log.Fatal("the databases type is unknown type")
	panic("db connect  has error")
}
