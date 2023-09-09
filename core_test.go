package ruxicore

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
)

func TestDBAuth(t *testing.T) {
	t.Run("TestDBAuth", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "test")
		t.Setenv("DB_HOST", "test")
		t.Setenv("DB_PORT", "test")
		t.Setenv("DB_USERNAME", "test")
		t.Setenv("DB_PASSWORD", "test")
		t.Setenv("DB", "test")

		dbAuth := GatherAuth()

		assert.Equal(t, "test", dbAuth.URL)
		assert.Equal(t, "test", dbAuth.Host)
		assert.Equal(t, "test", dbAuth.Port)
		assert.Equal(t, "test", dbAuth.Username)
		assert.Equal(t, "test", dbAuth.Password)
		assert.Equal(t, "test", dbAuth.Name)
	})
}

func TestHealthCheck(t *testing.T) {
	t.Run("TestHealthCheck", func(t *testing.T) {
		context := GetTestGinContext()
		HealthCheck(context)
		status := context.Writer.Status()
		if status != 200 {
			t.Errorf("Got %d, want %d", status, 200)
		}
	})
}

func TestInitDB(t *testing.T) {
	t.Run("TestInitDB", func(t *testing.T) {
		mockDB, _, _ := sqlmock.New()
		dialector := postgres.New(postgres.Config{
			Conn:       mockDB,
			DriverName: "postgres",
		})
		db, _ := InitDB(dialector)

		if db.DB == nil {
			t.Error("db.DB is nill")
		}
	})
	t.Run("TestInitDB Fail", func(t *testing.T) {
		_, err := InitDB(nil)

		if err != nil {
			t.Error("InitDB err should not be nil")
		}
	})
}

func TestRuxiGin(t *testing.T) {
	RuxiGin()
}

// Test that the GetLogger function returns a logger with default settings when called with a valid app name
func TestGetLogger_ReturnsLoggerWithDefaultSettings(t *testing.T) {
	// Set up test environment
	os.Setenv("LOG_LEVEL", "255")
	os.Setenv("LOG_SERVER", "")

	// Call the GetLogger function
	logger := GetLogger("myApp")

	// Assert that the logger has the expected default settings
	assert.Equal(t, zerolog.DebugLevel, logger.GetLevel())
	assert.Equal(t, time.RFC3339Nano, zerolog.TimeFieldFormat)
}

func CheckStringContains(t *testing.T, a, b string) {
	t.Helper()
	if !strings.Contains(a, b) {
		t.Errorf("Got: %s \n Want: %s\n", a, b)
	}
}

func GetTestGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	u, _ := url.Parse("https://ruxi.dev/test")
	ctx.Request = &http.Request{
		Header: make(http.Header),
		Method: "TEST",
		URL:    u,
		Proto:  "test",
	}
	ctx.Request.Header.Add("User-Agent", "test")
	return ctx
}
