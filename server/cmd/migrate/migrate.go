package main

import (
	"candly/internal/config"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"net/url"
	"os"
	"strconv"
)

func main() {
	c := config.GetConfig()

	m, err := migrate.New(
		"file://internal/db/migrations",
		"postgresql://"+c.Db.Username+":"+url.QueryEscape(c.Db.Password)+"@"+c.Db.Host+"/"+c.Db.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	if len(os.Args) == 3 {
		args := os.Args[1:]
		version, err := strconv.Atoi(args[1])

		if err != nil {
			fmt.Println("Invalid number")
			return
		}
		if args[0] == "-f" {

			if err = m.Force(version); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to force: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if args[0] == "-u" {

			if err = m.Steps(version); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to up: %v\n", err)
				os.Exit(1)
			}
			return
		}
		if args[0] == "-d" {

			if err = m.Steps(-version); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to force: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

}
