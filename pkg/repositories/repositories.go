package repositories

import (
	"allie/pkg/repositories/mage"

	"go.uber.org/fx"
)

var Module = fx.Options(
	mage.Module,
)
