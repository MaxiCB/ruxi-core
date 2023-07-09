package ruxicore

import (
	"context"
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ErrorLogger   *log.Logger
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
)

func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}

type DBAuth struct {
	URL      string
	Host     string
	Port     string
	Username string
	Password string
	Name     string
}

type DB struct {
	DB      *gorm.DB
	Context context.Context
}

func GatherAuth() *DBAuth {
	dbAuth := DBAuth{
		os.Getenv("DATABASE_URL"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_USERNAME"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_NAME"),
	}
	return &dbAuth
}

func InitDB(app_name string) *DB {
	//dbAuth := GatherAuth()
	dsn := "host=ruxi-backend user=postgres password=password dbname=postgres port=5432 sslmode=disable"
	fmt.Print(dsn)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		},
	)
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	db := DB{}
	db.DB = gdb
	db.Context = context.Background()
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
	f, _ := os.Create(fmt.Sprintf("%s/ruxi-gin.log", os.TempDir()))
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
	f, _ := os.Create(fmt.Sprintf("%s/%s.log", os.TempDir(), service_name))
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
