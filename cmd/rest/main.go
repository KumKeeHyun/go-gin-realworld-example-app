package main

import (
	"fmt"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/pkg/jwtutil"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"moul.io/zapgorm2"
)

func main() {
	config, err := readConfig()
	if err != nil {
		panic(err)
	}
	logger, err := InitZapLogger(config)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	logger.Sugar().Infow("read config", "config", config)

	r, err := InitRouter(config, logger)
	if err != nil {
		panic(err)
	}

	addr := config.Server.Host + ":" + config.Server.Port
	if config.Server.CertFile == "" || config.Server.KeyFile == "" {
		log.Fatal(r.Run(addr))
	} else {
		log.Fatal(r.RunTLS(addr, config.Server.CertFile, config.Server.KeyFile))
	}
}

func InitRouter(config *config, logger *zap.Logger) (*gin.Engine, error) {
	switch config.Datasource.DBType {
	case "sqlite":
		return InitRouterUsingSqlite(config, logger)
	case "postgres":
		return InitRouterUsingPostgres(config, logger)
	default:
		return nil, fmt.Errorf("invalid dbType: %s", config.Datasource.DBType)
	}
}

func InitZapLogger(config *config) (*zap.Logger, error) {
	switch config.Logger.Profile {
	case "dev":
		return zap.NewDevelopment()
	case "prod":
		return zap.NewProduction()
	default:
		return nil, fmt.Errorf("invalid logger profile: %s", config.Logger.Profile)
	}
}

func InitDatasource(config *config, logger *zap.Logger) (db *gorm.DB, err error) {
	gormLogger := zapgorm2.New(logger)
	gormLogger.SetAsDefault()
	gormCfg := &gorm.Config{
		Logger: gormLogger,
	}

	switch config.Datasource.DBType {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.Datasource.Url), gormCfg)
	case "postgres":
		db, err = gorm.Open(postgres.Open(config.Datasource.PostgresConfig), gormCfg)
	default:
		return nil, fmt.Errorf("invalid dbType: %s", config.Datasource.DBType)
	}
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Follow{},
		&domain.Article{},
		&domain.Favorite{},
		&domain.Comment{},
	)
	return
}

func InitJwtUtil(config *config) *jwtutil.JwtUtil {
	return jwtutil.New(jwt.SigningMethodHS256, []byte(config.Jwt.SecretKey))
}
