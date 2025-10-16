package main

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	//go:embed res/*
	resources  embed.FS
	Conf       = newConfig()
	gormDB     *gorm.DB
	sqlDB      *sql.DB
	adminEmail = "admin@admin.invalid"
)

func main() {
	var (
		e   = echo.New()
		err error
	)
	//GORM
	if gormDB, err = gorm.Open(postgres.New(postgres.Config{DSN: Conf.DatabaseUrl}),
		&gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Info)}); err != nil {
		logrus.Fatal(err)
	}
	if sqlDB, err = gormDB.DB(); err != nil {
		logrus.Fatal(err)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	e.GET("/reset/:token", resetTestDb)
	//reset on init
	if err = resetTestDbImpl(Conf.ResetToken); err != nil {
		logrus.Fatal(err)
	}
	if err = e.Start(Conf.ServerHost + ":" + strconv.Itoa(Conf.ServerPort)); err != nil {
		logrus.Fatal(err)
	}
}

func resetTestDb(e echo.Context) error {
	var (
		token = e.Param("token")
	)
	return resetTestDbImpl(token)
}

func resetTestDbImpl(token string) error {
	if token != Conf.ResetToken {
		logrus.Fatal("Reset token incorrect, expecting {}", Conf.ResetToken)
		return errors.New("reset token incorrect")
	}
	if strings.Contains(Conf.DatabaseUrl, "blaze_test") {
		ctx := context.Background()
		//the order of removal here matters, because there are no cascades, and we can get foreign key constraint violation
		gormDB.Table("zones").Where("1 = 1").Delete(ctx)
		gormDB.Table("banners").Where("1 = 1").Delete(ctx)
		gormDB.Table("files").Where("1 = 1").Delete(ctx)
		gormDB.Table("sites").Where("1 = 1").Delete(ctx)
		gormDB.Table("campaigns").Where("1 = 1").Delete(ctx)
		gormDB.Table("users").Where("email != ?", adminEmail).Delete(ctx)
		return nil
	} else {
		logrus.Fatal("Trying to reset a non-test db? {}", Conf.DatabaseUrl)
		return errors.New("trying to reset a non-test db")
	}
}

// Generates a password hash for the 'admin' password.
// This is the default password for a new installation. See migrations/000002_create_user_admin.up.sql
func genAdminCredentials() {
	var (
		password = "admin"
	)
	println(password)
	passwordHashBbs, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	println(string(passwordHashBbs))
	if bcrypt.CompareHashAndPassword(passwordHashBbs, []byte(password)) != nil {
		println("Password hash does not match")
	}
}

// Generates 000005_insert_content_categories.up.sql with inserts for content_category.
func genCategories() {
	var (
		tableName      = " content_category "
		columns        = "(id,parent_id,name,tier_1,tier_2,tier_3,tier_4)"
		values         = "(%s,%s,%s,%s,%s,%s,%s)"
		insertTemplate = "insert into" + tableName + columns + " values " + values + ";"
		buffer         bytes.Buffer
	)

	inReader := csv.NewReader(bytes.NewReader(GetFileBytes("res/Content Taxonomy 3.1.tsv", &resources)))
	inReader.Comma = '\t'
	records, err := inReader.ReadAll()
	if err != nil {
		println(err.Error())
	}
	records = records[2:] // reslice to omit header
	for _, record := range records {
		if !IsDigitsOnly(record[0]) {
			record[0] = strconv.Itoa(Hash(record[0]))
		}
		if !IsDigitsOnly(record[1]) {
			record[1] = strconv.Itoa(Hash(record[1]))
		}
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i][1] == "" {
			return true
		}
		if records[j][1] == "" {
			return false
		}
		return First(strconv.Atoi(records[i][0])) < First(strconv.Atoi(records[j][0]))
	})
	for _, record := range records {
		buffer.WriteString(fmt.Sprintf(insertTemplate,
			If(record[0] == "", "null", record[0]),
			If(record[1] == "", "null", record[1]),
			If(record[2] == "", "null", "'"+strings.ReplaceAll(record[2], "'", "''")+"'"),
			If(record[3] == "", "null", "'"+strings.ReplaceAll(record[3], "'", "''")+"'"),
			If(record[4] == "", "null", "'"+strings.ReplaceAll(record[4], "'", "''")+"'"),
			If(record[5] == "", "null", "'"+strings.ReplaceAll(record[5], "'", "''")+"'"),
			If(record[6] == "", "null", "'"+strings.ReplaceAll(record[6], "'", "''")+"'"),
		))
		buffer.WriteString("\n")
	}
	err = os.WriteFile("000005_insert_content_categories.up.sql", buffer.Bytes(), 0644)
	if err != nil {
		println(err.Error())
	}
}
