package ruxicore

import (
	"context"
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type DB struct {
	connection *pgdriver.Connector
	sqlDB      *sql.DB
	bunDB      *bun.DB
	context    context.Context
}

func InitDB(app_name string) *DB {
	db := DB{}
	db.connection = pgdriver.NewConnector(
		pgdriver.WithAddr(os.Getenv("DB_HOST")+":"+os.Getenv("DB_PORT")),
		pgdriver.WithUser(os.Getenv("DB_USERNAME")),
		pgdriver.WithPassword(os.Getenv("DB_PASSWORD")),
		pgdriver.WithDatabase(os.Getenv("DB")),
		pgdriver.WithApplicationName(app_name),
		pgdriver.WithInsecure(true),
	)
	db.sqlDB = sql.OpenDB(db.connection)
	db.bunDB = bun.NewDB(db.sqlDB, pgdialect.New())
	db.context = context.Background()
	return &db
}

func verifySession(options *sessmodels.VerifySessionOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		session.VerifySession(options, func(rw http.ResponseWriter, r *http.Request) {
			c.Request = c.Request.WithContext(r.Context())
			c.Next()
		})(c.Writer, c.Request)
		c.Abort()
	}
}