package httpx

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/ccfos/nightingale/v6/pkg/aop"
	"github.com/ccfos/nightingale/v6/pkg/version"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Host             string
	Port             int
	BasePath         string
	CertFile         string
	KeyFile          string
	PProf            bool
	PrintAccessLog   bool
	PrintBody        bool
	ExposeMetrics    bool
	ShutdownTimeout  int
	MaxContentLength int64
	ReadTimeout      int
	WriteTimeout     int
	IdleTimeout      int
	JWTAuth          JWTAuth
	ProxyAuth        ProxyAuth
	ShowCaptcha      ShowCaptcha
	APIForAgent      BasicAuths
	APIForService    BasicAuths
	RSA              RSAConfig
}

type RSAConfig struct {
	OpenRSA           bool
	RSAPublicKey      []byte
	RSAPublicKeyPath  string
	RSAPrivateKey     []byte
	RSAPrivateKeyPath string
	RSAPassWord       string
}

type ShowCaptcha struct {
	Enable bool
}

type BasicAuths struct {
	BasicAuth gin.Accounts
	Enable    bool
}

type ProxyAuth struct {
	Enable            bool
	HeaderUserNameKey string
	DefaultRoles      []string
}

type JWTAuth struct {
	SigningKey     string
	AccessExpired  int64
	RefreshExpired int64
	RedisKeyPrefix string
}

func GinEngine(mode string, cfg Config) *gin.Engine {
	gin.SetMode(mode)

	loggerMid := aop.Logger(aop.LoggerConfig{PrintBody: cfg.PrintBody})
	recoveryMid := aop.Recovery()

	if strings.ToLower(mode) == "release" {
		aop.DisableConsoleColor()
	}

	r := gin.New()

	//设置basePath,作为项目前缀
	setBasePath(r, cfg.BasePath)

	r.Use(recoveryMid)

	// whether print access log
	if cfg.PrintAccessLog {
		r.Use(loggerMid)
	}

	if cfg.PProf {
		pprof.Register(r, "/api/debug/pprof")
	}

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/pid", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%d", os.Getpid()))
	})

	r.GET("/ppid", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("%d", os.Getppid()))
	})

	r.GET("/addr", func(c *gin.Context) {
		c.String(200, c.Request.RemoteAddr)
	})

	r.GET("/api/n9e/version", func(c *gin.Context) {
		c.String(200, version.Version)
	})

	if cfg.ExposeMetrics {
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	return r
}

func Init(cfg Config, handler http.Handler) func() {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	go func() {
		fmt.Println("http server listening on:", addr)

		var err error
		if cfg.CertFile != "" && cfg.KeyFile != "" {
			srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			err = srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.ShutdownTimeout))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Println("cannot shutdown http server:", err)
		}

		select {
		case <-ctx.Done():
			fmt.Println("http exiting")
		default:
			fmt.Println("http server stopped")
		}
	}
}

// setBasePath 设置项目名前缀basePath,因为gin暂时不支持直接修改RouterGroup的basePath,使用unsafe.Pointer修改
// 需要在路由初始化前调用
func setBasePath(r *gin.Engine, basePath string) {
	if basePath == "" || basePath == "/" {
		return
	}

	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath = basePath + "/"
	}
	//因为Engine匿名注入了RouterGroup,所以直接获取Engine的反射对象
	engine := reflect.ValueOf(r).Elem()
	//获取RouterGroup的basePathValueOf属性反射值对象
	basePathValueOf := engine.FieldByName("basePath")
	//获取basePath的UnsafeAddr
	p := unsafe.Pointer(basePathValueOf.UnsafeAddr())
	//重新赋值basePath的反射值,NewAt默认返回的是指针,使用Elem获取反射值对象
	basePathValueOf = reflect.NewAt(basePathValueOf.Type(), p).Elem()
	//设置反射值
	basePathValueOf.Set(reflect.ValueOf(basePath))
}
