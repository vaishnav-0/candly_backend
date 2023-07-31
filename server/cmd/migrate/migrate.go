package main

import (
	"candly/internal/config"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"net/url"
	"os"
)

func main() {

	parser := argparse.NewParser("migrate", "go migrate for candly")

	verCmd := parser.NewCommand("v", "Get current version and dirty flag")

	downCmd := parser.NewCommand("down", "migrate all the way down")
	upCmd := parser.NewCommand("up", "migrate all the way up")

	stepCmd := parser.NewCommand("step", "step to version")
	ver := stepCmd.IntPositional(&argparse.Options{Help: "version", Required: true})

	forceCmd := parser.NewCommand("force", "force to version")
	fVer := forceCmd.IntPositional(&argparse.Options{Help: "version", Required: true})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	c := config.GetConfig()

	m, err := migrate.New(
		"file://internal/db/migrations",
		"postgresql://"+c.Db.Username+":"+url.QueryEscape(c.Db.Password)+"@"+c.Db.Host+"/"+c.Db.Name+"?sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	if forceCmd.Happened() {

		if err = m.Force(*fVer); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to force: %v\n", err)
			os.Exit(1)
		}
		return
	} else if upCmd.Happened() {

		err = m.Up()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to up: %v\n", err)
			os.Exit(1)
		}
	} else if stepCmd.Happened() {
		err = m.Steps(*ver)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to up: %v\n", err)
			os.Exit(1)
		}
	} else if downCmd.Happened() {

		err = m.Down()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to down: %v\n", err)
			os.Exit(1)
		}
	} else if verCmd.Happened() {
		fmt.Println(m.Version())
	}

}
