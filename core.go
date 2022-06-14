package ruxicore

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/supertokens"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

var (
	ErrorLogger   *log.Logger
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
)

type DB struct {
	connection *pgdriver.Connector
	SqlDB      *sql.DB
	BunDB      *bun.DB
	Context    context.Context
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
	db.SqlDB = sql.OpenDB(db.connection)
	db.BunDB = bun.NewDB(db.SqlDB, pgdialect.New())
	db.Context = context.Background()
	if os.Getenv("DB_LOGS") != "" {
		db.BunDB.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}
	if dbErr := db.BunDB.Ping(); dbErr != nil {
		panic(dbErr)
	}
	return &db
}

func VerifySession(options *sessmodels.VerifySessionOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		session.VerifySession(options, func(rw http.ResponseWriter, r *http.Request) {
			c.Request = c.Request.WithContext(r.Context())
			c.Next()
		})(c.Writer, c.Request)
		c.Abort()
	}
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

func RuxiLogger() gin.HandlerFunc {
	gin.DisableConsoleColor()
	f, _ := os.Create("/log/ruxi-gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

func InitLogger(service_name string) {
	f, _ := os.Create(fmt.Sprintf("/log/%s.log", service_name))
	ErrorLogger = log.New(io.MultiWriter(f, os.Stdout), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	InfoLogger = log.New(io.MultiWriter(f, os.Stdout), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(io.MultiWriter(f, os.Stdout), "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func LogInfo(err string, c *gin.Context) {
	if InfoLogger == nil {
		panic("InfoLogger not initialized")
	}
	if c != nil {
		request := c.Request
		request.ParseForm()
		request_info, _ := fmt.Printf("[%s] '%s' {%s} - %s",
			request.Method,
			request.URL,
			request.PostForm,
			err,
		)
		InfoLogger.Println(request_info)
	} else {
		InfoLogger.Println(err)
	}

}

func LogError(err string, c *gin.Context) {
	if ErrorLogger == nil {
		panic("ErrorLogger not initialized")
	}
	if c != nil {
		request := c.Request
		request.ParseForm()
		request_info, _ := fmt.Printf("[%s] '%s' {%s} - %s",
			request.Method,
			request.URL,
			request.PostForm,
			err,
		)
		ErrorLogger.Println(request_info)
	} else {
		ErrorLogger.Println(err)
	}

}

func LogWarning(err string, c *gin.Context) {
	if WarningLogger == nil {
		panic("WarningLogger not initialized")
	}
	if c != nil {
		request := c.Request
		request.ParseForm()
		request_info, _ := fmt.Printf("[%s] '%s' {%s} - %s",
			request.Method,
			request.URL,
			request.PostForm,
			err,
		)
		WarningLogger.Println(request_info)
	} else {
		WarningLogger.Println(err)
	}

}

func RuxiGin() *gin.Engine {
	router := gin.Default()
	router.Use(RuxiLogger())
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders: append([]string{"content-type"},
			supertokens.GetAllCORSHeaders()...),
		AllowCredentials: true,
	}))

	router.Use(func(c *gin.Context) {
		supertokens.Middleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
		c.Abort()
	})

	router.GET("/liveness", HealthCheck)
	return router
}
