package main

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func GetFileBytes(fileName string, fs *embed.FS) []byte {
	file, _ := fs.Open(fileName)
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(file); err != nil {
		println("Couldn't read from file: {}", err.Error())
		return nil
	}
	return buf.Bytes()
}

func First[T, U any](val T, _ U) T {
	return val
}

func Hash(s string) int {
	h := fnv.New32()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0
	}
	return int(h.Sum32()) % 2147483647 //max positive integer in PostgreSQL (4 bytes)
}

func IsDigitsOnly(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func If[T any](cond bool, vTrue, vFalse T) T {
	if cond {
		return vTrue
	}
	return vFalse
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
