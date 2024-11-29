package main

import (
	"flag"
	"log"

	"github.com/bennyscetbun/xxxyourappyyy/backend/database"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	flag.Parse()
	db, err := database.OpenPSQL()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = database.CheckUpAndDown(db)
	if err != nil {
		log.Fatal(err)
	}
	err = database.CheckEmpty(db)
	if err != nil {
		log.Fatal(err)
	}
}
