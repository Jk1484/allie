package main

import (
	"allie/internal/api/handlers"
	"allie/internal/api/server"
	"allie/internal/db"
	"allie/internal/repositories"
	"allie/internal/services"
	"allie/pkg/config"
	"allie/pkg/logger"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Options(
			config.Module,
			logger.Module,
			db.Module,
			repositories.Module,
			services.Module,
			handlers.Module,
			server.Module,
		),
	).Run()
}
