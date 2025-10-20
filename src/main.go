package main

import (
	"embed"
	"errors"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	//go:embed res/*
	resources embed.FS
	Conf      = newConfig()
)

func main() {
	var (
		e   = echo.New()
		err error
	)
	e.POST("/reset", reset)
	if err = e.Start(Conf.ServerHost + ":" + strconv.Itoa(Conf.ServerPort)); err != nil {
		logrus.Fatal(err)
	}
}

type ResetRequest struct {
	ResetToken  *string `json:"resetToken,omitempty"`
	DatabaseUrl *string `json:"databaseUrl,omitempty"`
	Sql         *string `json:"sql,omitempty"`
}

func reset(e echo.Context) error {
	var (
		resetRequest ResetRequest
		err          error
	)
	if err = e.Bind(&resetRequest); err != nil {
		logrus.Error(err.Error())
		return err
	}
	return resetImpl(&resetRequest)
}

func resetImpl(resetRequest *ResetRequest) error {
	var (
		gormDB *gorm.DB
		err    error
	)
	if *resetRequest.ResetToken != Conf.ResetToken {
		logrus.Fatal("Reset token incorrect, expecting {}", Conf.ResetToken)
		return errors.New("reset token incorrect")
	} else {
		//GORM
		if gormDB, err = gorm.Open(postgres.New(postgres.Config{DSN: *resetRequest.DatabaseUrl}),
			&gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Info)}); err != nil {
			logrus.Fatal(err)
			return err
		} else {
			splits := strings.Split(*resetRequest.Sql, ";")
			for i := range splits {
				gormDB.Exec(splits[i])
			}
		}
	}
	return nil
}
