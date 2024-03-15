package services

import (
	"allie/pkg/services/arena"
	"allie/pkg/services/mage"

	"go.uber.org/fx"
)

var Module = fx.Options(
	mage.Module,
	arena.Module,
)
