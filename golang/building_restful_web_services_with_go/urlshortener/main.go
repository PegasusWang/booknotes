package main

import (
	"log"

	"github.com/narenaryan/models"
)

func main() {
	db, err := models.InitDB()
	if err != nil {
		log.Println(db)
	}
}
