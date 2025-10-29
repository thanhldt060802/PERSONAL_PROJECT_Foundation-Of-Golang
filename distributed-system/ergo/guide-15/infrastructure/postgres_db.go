package infrastructure

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var PostgresDB *bun.DB

func InitPostgesDB() {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		"postgres",
		"12345678",
		"localhost",
		"5432",
		"my_db",
	)

	connection, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Connect to PostgreSQL failed: ", err)
	}

	if err := connection.Ping(); err != nil {
		log.Fatal("Ping to PostgreSQL failed: ", err)
	}

	PostgresDB = bun.NewDB(connection, pgdialect.New())

	if err := PostgresDB.Ping(); err != nil {
		log.Fatal("Ping to PostgreSQL with Bun failed: ", err)
	}

	log.Println("Connect to PostgreSQL with Bun successful")
}
