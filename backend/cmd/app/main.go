package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bennyscetbun/xxxyourappyyy/backend/api"
	"github.com/bennyscetbun/xxxyourappyyy/backend/database"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/domains"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/environment"
	"github.com/bennyscetbun/xxxyourappyyy/backend/internal/logger"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/ztrue/tracerr"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func openDB() (*gorm.DB, *sql.DB, error) {
	db, err := database.OpenPSQL()
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	err = database.MigratePSQL(db)
	if err != nil {
		db.Close()
		return nil, nil, tracerr.Wrap(err)
	}
	gormDb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		db.Close()
		return nil, nil, tracerr.Wrap(err)
	}
	return gormDb, db, nil
}

func multiplex(grpcServer *grpc.Server, wrappedGRPCServer *grpcweb.WrappedGrpcServer, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else if wrappedGRPCServer.IsGrpcWebRequest(r) {
			wrappedGRPCServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

const paramPathName = "p"

func main() {
	domainName := flag.String("domain", "", "set the domain, if not set https will be disabled")
	serveFromServer := flag.String("serve_bundle_url", "", "if set serve bundle from server")
	flag.Parse()
	tokenSecret := environment.MustGetenvString("TOKEN_SECRET", "")
	if tokenSecret == "" {
		logger.Fatalln("Set TOKEN_SECRET environment variable")
	}

	gormDB, db, err := openDB()
	if err != nil {
		log.Fatalln(tracerr.Sprint(err))
	}
	defer db.Close()
	grpcServer, err := api.CreateServer(gormDB, flag.Args()[2], []byte(tokenSecret))
	if err != nil {
		logger.Fatalln(tracerr.Sprint(err))
	}
	wrappedGrpc := grpcweb.WrapServer(grpcServer)
	ginHandler := gin.Default()
	if *serveFromServer == "" {
		ginHandler.Static("/assets", flag.Args()[1])
	} else {
		remote, err := url.Parse(*serveFromServer)
		if err != nil {
			panic(err)
		}
		ginHandler.GET("/assets/*filepath", func(ctx *gin.Context) {
			proxy := httputil.NewSingleHostReverseProxy(remote)
			//Define the director func
			//This is a good place to log, for example
			proxy.Director = func(req *http.Request) {
				req.Header = ctx.Request.Header
				req.Host = remote.Host
				req.URL.Scheme = remote.Scheme
				req.URL.Host = remote.Host
				req.URL.Path = ctx.Param("filepath")
			}
			proxy.ServeHTTP(ctx.Writer, ctx.Request)
			ctx.Abort()
		})
	}

	ginHandler.GET("/favicon.ico", func(c *gin.Context) {
		c.Request.URL.Path = "/assets/favicon.ico"
		ginHandler.HandleContext(c)
	})

	ginHandler.NoRoute(func(ctx *gin.Context) {
		http.ServeFile(ctx.Writer, ctx.Request, flag.Args()[0])
		ctx.Abort()
	})

	handler := multiplex(grpcServer, wrappedGrpc, ginHandler)

	srvHTTP := &http.Server{Addr: ":8080", Handler: handler}
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		if err := srvHTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln("listen http:", err)
		}
		wg.Done()
	}()
	runWithContextDone := make(chan struct{})

	handlerWithSubdomain := domains.NewDomain()

	if *domainName != "" {
		handlerWithSubdomain.GetOrCreateDomainsHandler(*domainName, func() http.Handler {
			return handler
		})
		wg.Add(1)
		go func() {
			if err := autotls.RunWithContext(ctx, handlerWithSubdomain, handlerWithSubdomain.GetDomains()...); err != nil && err != http.ErrServerClosed {
				log.Print("listen https:", err)
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(runWithContextDone)
	}()
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctxCancel()
	srvHTTP.Shutdown(context.Background())
	fmt.Println("gracefully quitting....")
	select {
	case <-runWithContextDone:
		fmt.Println("gracefully quitted")
	case <-time.After(10 * time.Second):
		fmt.Println("timeout quitted")
	}
}
