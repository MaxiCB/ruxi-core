package ruxicore

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
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

func GatherAuth() *DBAuth {
	// GatherAuth retrieves the authentication details for the database connection.
	// It retrieves the necessary information from environment variables and creates a DBAuth struct with the gathered values.

	// Retrieve the values of the database authentication details from the corresponding environment variables.
	dbAuth := DBAuth{}
	if url := os.Getenv("DATABASE_URL"); url != "" {
		dbAuth.URL = url
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		dbAuth.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		dbAuth.Port = port
	}
	if username := os.Getenv("DB_USERNAME"); username != "" {
		dbAuth.Username = username
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		dbAuth.Password = password
	}
	if db := os.Getenv("DB"); db != "" {
		dbAuth.Name = db
	}

	// Return a pointer to the created DBAuth struct.
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

func GetGitRevision() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, v := range buildInfo.Settings {
			if v.Key == "vcs.revision" {
				return v.Value
			}
		}
	}
	return ""
}

func GetGoVersion() string {
	return runtime.Version()
}

func GetLogger(appName string) zerolog.Logger {
	once.Do(func() {
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
		if err != nil {
			logLevel = zerolog.DebugLevel
		}

		var output io.Writer = zerolog.ConsoleWriter{Out: os.Stdout}
		log = zerolog.New(output).Level(zerolog.Level(logLevel)).With().Timestamp().Str("app", appName).Logger()

		log_server := os.Getenv("LOG_SERVER")
		if log_server != "" {
			server, err := net.ResolveTCPAddr("tcp", log_server)
			if err != nil {
				log.Fatal().Err(err).Msg("Unable to resolve LOG_SERVER address")
			}
			conn, err := net.DialTCP("tcp", nil, server)
			if err != nil {
				log.Fatal().Err(err).Msg("Unable to dial LOG_SERVER address")
			}
			output = zerolog.MultiLevelWriter(os.Stdout, conn)
		}

		log = zerolog.New(output).Level(zerolog.Level(logLevel)).With().Timestamp().Str("app", appName).Str("git_revision", GetGitRevision()).Str("go_version", GetGoVersion()).Logger()
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
