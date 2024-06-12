package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/danmcfan/video-microservice-golang/internal"
	"github.com/danmcfan/video-microservice-golang/internal/database"
)

func main() {
	db, err := sql.Open("sqlite3", "./temp/test.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
		return
	}

	err = database.CreateVideoTable(db)
	if err != nil {
		log.Fatalf("Failed to create video table: %v", err)
		return
	}

	err = database.CreateFrameTable(db)
	if err != nil {
		log.Fatalf("Failed to create frame table: %v", err)
		return
	}

	err = database.CreateFrameIndex(db)
	if err != nil {
		log.Fatalf("Failed to create frame index: %v", err)
		return
	}

	storageDir := "./temp/storage"

	if storageDir == "" {
		storageDir, err = os.MkdirTemp("", "")
		if err != nil {
			log.Fatalf("Failed to create storage directory: %v", err)
			return
		}
	} else {
		err = os.MkdirAll(storageDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create storage directory: %v", err)
			return
		}
	}

	router := internal.CreateRouter(db, storageDir)
	router.Run(":8080")

	db.Close()
}