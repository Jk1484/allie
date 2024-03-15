package main

import (
	"allie/configs"
	"allie/pkg/db"
	"allie/pkg/handlers"
	"allie/pkg/handlers/server"
	"allie/pkg/logger"
	"allie/pkg/repositories"
	"allie/pkg/services"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Options(
			configs.Module,
			logger.Module,
			db.Module,
			repositories.Module,
			services.Module,
			handlers.Module,
			server.Module,
		),
	).Run()
}
