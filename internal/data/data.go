package data

import (
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
var ProviderSet = wire.NewSet(NewData, NewReviewRepo, Connect)

// Data .
type Data struct {
	// TODO wrapped database client
	query *query.Query
	log   *log.Helper
}

// NewData .
func NewData(db *gorm.DB, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db)
	return &Data{
		query: query.Q,
		log:   log.NewHelper(logger),
	}, cleanup, nil
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
			log.Fatal("mysql db connect has error")
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
