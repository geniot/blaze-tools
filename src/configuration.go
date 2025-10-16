package main

import (
	"github.com/num30/config"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ServerHost  string `default:"localhost" validate:"required" envvar:"SERVER_HOST"`
	ServerPort  int    `default:"8444" validate:"required" envvar:"SERVER_PORT"`
	DatabaseUrl string `default:"postgres://blaze_test:blaze_test@localhost:9876/blaze_test" validate:"required" envvar:"DATABASE_URL"`
	ResetToken  string `default:"sometoken" validate:"required" envvar:"RESET_TOKEN"`
}

func newConfig() *Config {
	var (
		c   Config
		err error
	)
	if err = config.NewConfReader("myconf").Read(&c); err != nil {
		logrus.Fatal(err)
	}
	return &c
}
