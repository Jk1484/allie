package configs

import (
	"encoding/json"
	"os"
	"strings"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(readFile),
	fx.Provide(New),
)

var cfg *configs

type Configs interface {
	Peek() *configs
}

type configs struct {
	Database Database `json:"database"`
}

type Database struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Params struct {
	fx.In
}

func New(p Params) Configs {
	return &configs{}
}

func (c *configs) Peek() *configs {
	return cfg
}

func readFile() *configs {
	wd, _ := os.Getwd()
	index := strings.LastIndex(wd, "allie")
	wd = wd[:index]

	f, err := os.Open(wd + "allie/configs/configs.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
