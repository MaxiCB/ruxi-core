package ruxicore

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
)

func TestIf(t *testing.T) {
	t.Run("Test If True", func(t *testing.T) {
		cond := If[bool]("test" == "test", true, false)
		if !cond {
			t.Errorf("Expected true")
		}
	})
	t.Run("Test If False", func(t *testing.T) {
		cond := If[bool]("test" == "not", true, false)
		if cond {
			t.Errorf("Expected true")
		}
	})
}

func TestDBAuth(t *testing.T) {
	t.Run("TestDBAuth", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "test")
		t.Setenv("DB_HOST", "test")
		t.Setenv("DB_PORT", "test")
		t.Setenv("DB_USERNAME", "test")
		t.Setenv("DB_PASSWORD", "test")
		t.Setenv("DB", "test")

		dbAuth := GatherAuth()

		CheckStringContains(t, dbAuth.URL, "test")
		CheckStringContains(t, dbAuth.Host, "test")
		CheckStringContains(t, dbAuth.Port, "test")
		CheckStringContains(t, dbAuth.Username, "test")
		CheckStringContains(t, dbAuth.Password, "test")
		CheckStringContains(t, dbAuth.Name, "test")
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
			Conn: mockDB, DriverName: "postgres",
		})
		db, _ := InitDB(dialector)
		if db.DB == nil {
			t.Error("db.DB is nill")
		}
	})
	t.Run("TestInitDB Fail", func(t *testing.T) {
		db, err := InitDB(nil)
		fmt.Printf("db: %v\n", db)
		fmt.Printf("err: %v\n", err)

		if err != nil {
			t.Error("InitDB err should not be nil")
		}
	})
}

func TestGetLogger(t *testing.T) {
	t.Run("Test GetLogger", func(t *testing.T) {
		logger := GetLogger("test")
		if zerolog.TimeFieldFormat != time.RFC3339Nano {
			t.Error("Logger TimeFieldFormat should be RFC3339Nano")
		}
		if logger.GetLevel() != zerolog.InfoLevel {
			t.Error("Logger level should be InfoLevel")
		}
	})
}

func TestRuxiGin(t *testing.T) {
	RuxiGin()
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
