package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/shared/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type GatewayConfig struct {
	Host           string
	Port           int
	GRPCAddress    string
	AllowedOrigins []string
}

type Gateway struct {
	cfg    *GatewayConfig
	server *http.Server
	mux    *runtime.ServeMux
	logger *logger.Logger
}

func NewGateway(cfg *GatewayConfig, log *logger.Logger) (*Gateway, error) {
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   false,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	g := &Gateway{
		cfg:    cfg,
		mux:    mux,
		logger: log,
	}

	return g, nil
}

func (g *Gateway) RegisterServices(ctx context.Context) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := pb.RegisterSessionServiceHandlerFromEndpoint(ctx, g.mux, g.cfg.GRPCAddress, opts); err != nil {
		return fmt.Errorf("failed to register session service: %w", err)
	}

	if err := pb.RegisterSchemaServiceHandlerFromEndpoint(ctx, g.mux, g.cfg.GRPCAddress, opts); err != nil {
		return fmt.Errorf("failed to register schema service: %w", err)
	}

	if err := pb.RegisterDataServiceHandlerFromEndpoint(ctx, g.mux, g.cfg.GRPCAddress, opts); err != nil {
		return fmt.Errorf("failed to register data service: %w", err)
	}

	g.logger.With().Str("grpc_address", g.cfg.GRPCAddress).Logger().Info("registered gRPC gateway services")

	return nil
}

func (g *Gateway) Start() error {
	handler := g.corsMiddleware(g.mux)

	addr := fmt.Sprintf("%s:%d", g.cfg.Host, g.cfg.Port)

	g.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	g.logger.With().Str("address", addr).Logger().Info("HTTP gateway listening")

	if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (g *Gateway) Stop(ctx context.Context) error {
	g.logger.Info("stopping HTTP gateway")

	if g.server == nil {
		return nil
	}

	if err := g.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	g.logger.Info("HTTP gateway stopped")
	return nil
}

func (g *Gateway) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if g.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(g.cfg.AllowedOrigins) == 0 || g.cfg.AllowedOrigins[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (g *Gateway) isOriginAllowed(origin string) bool {
	if len(g.cfg.AllowedOrigins) == 0 {
		return true
	}

	for _, allowed := range g.cfg.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}

func customHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Authorization":
		return key, true
	case "Content-Type":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
