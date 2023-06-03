package main

import (
	"os"
)

func processError(err error) {
	ErrorLogger.Println(err)
	os.Exit(2)
}
