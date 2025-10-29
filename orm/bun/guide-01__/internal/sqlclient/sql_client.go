package sqlclient

import (
	"database/sql"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var SqlClientConnInstance ISqlClientConn

type ISqlClientConn interface {
	GetDB() *bun.DB
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
	DB *bun.DB
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
	postgresConn := pgdriver.NewConnector(
		pgdriver.WithAddr(fmt.Sprintf("%v:%v", c.Host, c.Port)),
		pgdriver.WithDatabase(c.Database),
		pgdriver.WithUser(c.Username),
		pgdriver.WithPassword(c.Password),
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithInsecure(true),
	)

	db := bun.NewDB(sql.OpenDB(postgresConn), pgdialect.New(), bun.WithDiscardUnknownColumns())
	if err := db.Ping(); err != nil {
		return err
	}
	c.DB = db

	return nil
}

func (c *SqlClientConn) GetDB() *bun.DB {
	return c.DB
}
