package app

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/skamenetskiy/grpc-app-deps/log"
	"google.golang.org/grpc"
)

type App interface {
	Start()

	HTTP() *http.Server
	GRPC() *grpc.Server
	Gateway() *runtime.ServeMux
	Router() chi.Router
}

func New(ctx context.Context, options ...Option) App {
	a := &app{
		ctx:        ctx,
		httpPort:   8080,
		grpcPort:   9000,
		router:     chi.NewRouter(),
		gatewayMux: runtime.NewServeMux(),
	}

	for _, opt := range options {
		opt(a)
	}

	if len(a.swaggerFile) > 0 {
		a.router.HandleFunc("/docs/api.swagger.json", a.serveSwaggerFile)
		a.router.HandleFunc("/docs/*", createDocsHandler())
	}

	a.router.Handle("/*", a.gatewayMux)

	a.grpcServer = grpc.NewServer()
	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.httpPort),
		Handler: a.router,
	}

	return a
}

type app struct {
	ctx context.Context

	httpPort int
	grpcPort int

	httpServer *http.Server
	grpcServer *grpc.Server
	gatewayMux *runtime.ServeMux

	router chi.Router

	swaggerFile []byte
}

func (a *app) Start() {
	shutdown := make(chan struct{}, 1)

	// start grpc server
	go func() {
		addr := fmt.Sprintf(":%d", a.grpcPort)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal(a.ctx, "failed to listen grpc", "error", err)
		}
		if err = a.grpcServer.Serve(listener); err != nil {
			log.Fatal(a.ctx, "failed to serve grpc", "error", err)
		}
	}()

	// start http server
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(a.ctx, "failed to listen http", "error", err)
		}
	}()

	// listen to shut down signal
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill)

		sig := <-ch
		log.Info(a.ctx, "shutdown gracefully", "signal", sig)

		ctx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
		defer cancel()

		a.grpcServer.GracefulStop()
		if err := a.httpServer.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown http server", "error", err)
		}

		shutdown <- struct{}{}
	}()

	<-shutdown
}

func (a *app) HTTP() *http.Server {
	return a.httpServer
}

func (a *app) GRPC() *grpc.Server {
	return a.grpcServer
}

func (a *app) Gateway() *runtime.ServeMux {
	return a.gatewayMux
}

func (a *app) Router() chi.Router {
	return a.router
}

func (a *app) serveSwaggerFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(a.swaggerFile)
}

//go:embed docs
var docsFS embed.FS

func createDocsHandler() http.HandlerFunc {
	fileServer := http.FileServer(http.FS(docsFS))
	return fileServer.ServeHTTP
}
