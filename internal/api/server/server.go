package server

import (
	"allie/internal/api/handlers"
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

var Module = fx.Options(fx.Invoke(Init))

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Handlers  handlers.Handlers
}

func Init(p Params) {
	mux := mux.NewRouter()

	mux.HandleFunc("/ws", http.HandlerFunc(p.Handlers.HandleWebsocket)).Methods("GET")

	server := http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	p.Lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go server.ListenAndServe()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return server.Shutdown(ctx)
			},
		},
	)
}
