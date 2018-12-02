package app

import (
    "fmt"
    "net/http"
    "go.uber.org/zap"
    "github.com/gin-gonic/gin"
    "github.com/spf13/viper"
    "github.com/pkg/errors"
    "github.com/zhanben/go_site/tool/db"
)

type Server struct {
    HttpServer  *http.Server
	u  *User
	ct *Comment

	logger *zap.Logger
	dbConn *db.DbConn
}

func NewServer(dbConn *db.DbConn, logger *zap.Logger)(*Server, error){
    var err error
    s := &Server{logger:logger}

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

    s.u.initRouter(rg)
    s.ct.initRouter(rg)
    return r
}

type stackTracer interface {
    StackTrace() errors.StackTrace
}

func abortWithError(l *zap.Logger, c *gin.Context, e error) {
    err, ok := errors.Cause(e).(stackTracer)
    if ok {
        st := err.StackTrace()
        l.Info(fmt.Sprintf("error: %s\n%+v", err, st[0:2]))
    } else {
        l.Info(fmt.Sprintf("%+v", e))
    }

    c.Abort()
    c.String(http.StatusBadGateway, "%+v", e)
}


