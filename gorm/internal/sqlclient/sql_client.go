package sqlclient

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var SqlClientConnInstance ISqlClientConn

type ISqlClientConn interface {
	GetDB() *gorm.DB
}

type SqlConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type SqlClientConn struct {
	SqlConfig
	DB *gorm.DB
}

func NewSqlClient(config SqlConfig) *SqlClientConn {
	client := &SqlClientConn{}
	client.SqlConfig = config
	if err := client.Connect(); err != nil {
		log.Fatalf("Connect to postgres failed: %v", err.Error())
	}
	return client
}

func (c *SqlClientConn) Connect() error {
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
		c.Host,
		c.Username,
		c.Password,
		c.Database,
		c.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	c.DB = db

	return nil
}

func (c *SqlClientConn) GetDB() *gorm.DB {
	return c.DB
}
