package repositories

import (
	"allie/internal/repositories/mage"

	"go.uber.org/fx"
)

var Module = fx.Options(
	mage.Module,
)
