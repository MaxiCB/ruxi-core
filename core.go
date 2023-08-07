package ruxicore

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/supertokens"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
type LogLevel struct {
	slug string
}

type LogMessage struct {
	LogType    LogLevel
	Message    string
	GinContext *gin.Context //optional
}

var (
	ERROR   = LogLevel{"ERROR"}
	WARNING = LogLevel{"WARNING"}
	INFO    = LogLevel{"INFO"}
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

func GatherAuth() *DBAuth {
	dbAuth := DBAuth{
		os.Getenv("DATABASE_URL"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB"),
	}
	return &dbAuth
}

func InitDB(dialector gorm.Dialector) (*DB, error) {
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
	gdb, err := gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		fmt.Fprintf(ErrorLogger.Writer(), "Unable to connect to database: %v\n", err)
		return nil, err
	}
	db := DB{}
	db.DB = gdb
	db.Context = context.Background()
	return &db, nil
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

func RuxiLogFormatter(param gin.LogFormatterParams) string {
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
}

func RuxiLogger(writers ...io.Writer) gin.HandlerFunc {
	gin.DefaultWriter = io.MultiWriter(writers...)
	return gin.LoggerWithFormatter(RuxiLogFormatter)
}

func InitLogger(service_name string, writers ...io.Writer) {
	ErrorLogger = log.New(io.MultiWriter(writers...), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	InfoLogger = log.New(io.MultiWriter(writers...), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(io.MultiWriter(writers...), "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Log(message LogMessage) {
	var logMessage string
	if message.GinContext != nil {
		request := message.GinContext.Request
		request.ParseForm()
		logMessage = fmt.Sprintf("[%s] '%s' {%s} - %s",
			request.Method,
			request.URL,
			request.PostForm,
			message.Message,
		)
	} else {
		logMessage = message.Message
	}

	switch message.LogType.slug {
	case ERROR.slug:
		ErrorLogger.Println(logMessage)
	case INFO.slug:
		InfoLogger.Println(logMessage)
	case WARNING.slug:
		WarningLogger.Println(logMessage)
	}
}

func GatherCorsAllowHeaders() []string {
	allowHeaders := []string{"content-type"}
	env := os.Getenv("production")
	if strings.EqualFold("production", env) {
		allowHeaders = append(allowHeaders, supertokens.GetAllCORSHeaders()...)
	}
	return allowHeaders
}

func AddSuperTokens(router *gin.Engine) {
	env := os.Getenv("production")
	if strings.EqualFold("production", env) {
		router.Use(func(c *gin.Context) {
			supertokens.Middleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				c.Next()
			})).ServeHTTP(c.Writer, c.Request)
			c.Abort()
		})
	}
}

func RuxiGin() *gin.Engine {
	router := gin.Default()
	gin.DisableConsoleColor()
	router.Use(RuxiLogger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders:     GatherCorsAllowHeaders(),
		AllowCredentials: true,
	}))

	router.GET("/liveness", HealthCheck)
	return router
}
