package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"github.com/zhanben/go_site/tool/db"
	"github.com/zhanben/go_site/tool/log"
	"github.com/zhanben/go_site/docs"
)

type Server struct {
	HttpServer *http.Server
	u          *User
	ct         *Comment

	logger *zap.SugaredLogger
	dbConn *db.DbConn
}

func NewServer(dbConn *db.DbConn, logger *zap.SugaredLogger) (*Server, error) {
	var err error
	s := &Server{logger: logger}

	s.u, err = NewUser(dbConn, logger)
	if err != nil {
		logger.Panic("init user failed")
	}

	s.ct, err = NewComment(dbConn, logger)
	if err != nil {
		logger.Panic("init Comment failed")
	}

	port := viper.GetInt("Server.Port")
	s.HttpServer = &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        s.initRouter(),
		MaxHeaderBytes: 1 << 20,
	}
	return s, nil

}

func (s *Server) initRouter() http.Handler {
	r := gin.Default()
	r.Use(gin.ErrorLogger())
	rg := r.Group("/api")

	rg.GET("echo", func(c *gin.Context) {
		c.JSON(200, "hello world!")
	})
	port := viper.GetInt("Server.Port")
	host := viper.GetString("Server.Host")
	s.u.initRouter(rg)
	s.ct.initRouter(rg)
	docs.SwaggerInfo.Title = "go-siteExample API"
	docs.SwaggerInfo.Description = "This is a sample server go-site server."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", host, port)
	docs.SwaggerInfo.BasePath = ""
	log.Logger.Infof("swagger info:%+v", docs.SwaggerInfo)
	rg.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func abortWithError(l *zap.SugaredLogger, c *gin.Context, e error) {
	err, ok := errors.Cause(e).(stackTracer)
	if ok {
		st := err.StackTrace()
		l.Info("error: %s\n%+v", err, st[0:2])
	} else {
		l.Info("%+v", e)
	}

	c.Abort()
	c.String(http.StatusBadGateway, "%+v", e)
}
