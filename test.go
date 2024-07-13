package main

import (
	"log"
	"os"
)

func init() {

	log.Printf("GOOGLE_CLIENT_ID: %s", os.Getenv("GOOGLE_CLIENT_ID"))
	log.Printf("GOOGLE_CLIENT_SECRET: %s", os.Getenv("GOOGLE_CLIENT_SECRET"))
}
