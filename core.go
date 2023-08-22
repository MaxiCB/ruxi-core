package ruxicore

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var once sync.Once
var log zerolog.Logger

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
	gdb, err := gorm.Open(dialector)
	if err != nil {
		return nil, err
	}
	db := DB{}
	db.DB = gdb
	db.Context = context.Background()
	return &db, nil
}

func GetLogger(appName string) zerolog.Logger {
	once.Do(func() {
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
		if err != nil {
			logLevel = int(zerolog.InfoLevel)
		}

		var output io.Writer = zerolog.ConsoleWriter{Out: os.Stdout}

		log_server := os.Getenv("LOG_SERVER")
		if log_server != "" {
			server, err := net.ResolveTCPAddr("tcp", log_server)
			if err != nil {
				panic("Unable to resolve LOG_SERVER address")
			}
			conn, err := net.DialTCP("tcp", nil, server)
			if err != nil {
				panic("Unable to connect to LOG_SERVER address")
			}
			output = zerolog.MultiLevelWriter(os.Stdout, conn)
		}

		var gitRevision string

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "vcs.revision" {
					gitRevision = v.Value
					break
				}
			}
		}

		log = zerolog.New(output).Level(zerolog.Level(logLevel)).With().Timestamp().Str("app", appName).Str("git_revision", gitRevision).Str("go_version", buildInfo.GoVersion).Logger()
	})

	return log
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

func RuxiGin() *gin.Engine {
	router := gin.Default()
	gin.DisableConsoleColor()

	router.GET("/liveness", HealthCheck)
	return router
}
