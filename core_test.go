package ruxicore

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
)

var logBuffer = bytes.Buffer{}

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

func TestLogger(t *testing.T) {
	t.Run("Log Error Fails", func(t *testing.T) {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("LogError should have panicked!")
				}
			}()
			Log(LogMessage{ERROR, "testing", nil})
		}()
	})

	t.Run("Log Info Fails", func(t *testing.T) {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("LogInfo should have panicked!")
				}
			}()
			Log(LogMessage{INFO, "testing", nil})
		}()
	})

	t.Run("Log Warning Fails", func(t *testing.T) {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("LogWarning should have panicked!")
				}
			}()
			Log(LogMessage{WARNING, "testing", nil})
		}()
	})

	t.Run("Logger Init", func(t *testing.T) {
		InitLogger("testing")

		loggers := []*log.Logger{ErrorLogger, InfoLogger, WarningLogger}

		for _, logger := range loggers {
			if logger == nil {
				t.Fatalf("%s logger nil", logger.Prefix())
			}
		}
	})

	t.Run("LogInfo no context", func(t *testing.T) {
		InfoLogger.SetOutput(&logBuffer)
		Log(LogMessage{INFO, "testing", nil})

		want := "testing"
		got := logBuffer.String()
		logBuffer.Reset()

		CheckStringContains(t, got, want)
	})

	t.Run("LogInfo context", func(t *testing.T) {
		ctx := GetTestGinContext()
		InfoLogger.SetOutput(&logBuffer)
		Log(LogMessage{INFO, "testing", ctx})

		want := "testing"
		got := logBuffer.String()
		logBuffer.Reset()

		CheckStringContains(t, got, want)
	})

	t.Run("LogWarning no context", func(t *testing.T) {
		WarningLogger.SetOutput(&logBuffer)
		Log(LogMessage{WARNING, "testing", nil})

		want := "testing"
		got := logBuffer.String()
		logBuffer.Reset()

		CheckStringContains(t, got, want)
	})

	t.Run("LogWarning context", func(t *testing.T) {
		ctx := GetTestGinContext()
		WarningLogger.SetOutput(&logBuffer)
		Log(LogMessage{WARNING, "testing", ctx})

		want := "testing"
		got := logBuffer.String()
		logBuffer.Reset()

		CheckStringContains(t, got, want)
	})

	t.Run("LogError no context", func(t *testing.T) {
		ErrorLogger.SetOutput(&logBuffer)
		Log(LogMessage{LogType: ERROR, Message: "testing"})

		want := "testing"
		got := logBuffer.String()
		logBuffer.Reset()

		CheckStringContains(t, got, want)
	})

	t.Run("LogError context", func(t *testing.T) {
		ctx := GetTestGinContext()
		ErrorLogger.SetOutput(&logBuffer)
		Log(LogMessage{LogType: ERROR, Message: "testing", GinContext: ctx})

		want := "testing"
		got := logBuffer.String()
		logBuffer.Reset()

		CheckStringContains(t, got, want)
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
		db, err := InitDB(nil)

		if db == nil {
			t.Error("InitDB db should be nil")
		}

		if err != nil {
			t.Error("InitDB err should not be nil")
		}
	})
}

func TestRuxiLogger(t *testing.T) {
	ctx := GetTestGinContext()
	RuxiLogger(&logBuffer)
	RuxiLogFormatter(gin.LogFormatterParams{
		TimeStamp:    time.Now(),
		Method:       "POST",
		Path:         "/test",
		Request:      ctx.Request,
		StatusCode:   200,
		Latency:      time.Duration(10 * time.Millisecond),
		ErrorMessage: "",
	})
}

func TestRuxiGin(t *testing.T) {
	RuxiGin()
}

func TestVerifySession(t *testing.T) {
	VerifySession(nil)
}

func ShouldPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() { _ = recover() }()
	f()
	t.Error("Should have panicked")
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
