package services

import (
	"allie/internal/services/arena"
	"allie/internal/services/mage"

	"go.uber.org/fx"
)

var Module = fx.Options(
	mage.Module,
	arena.Module,
)
