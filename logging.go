package main

import (
	"log"
	"os"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func init() {

	flags := log.Ldate | log.Ltime

	InfoLogger = log.New(os.Stdout, "INFO ", flags)
	WarningLogger = log.New(os.Stdout, "WARNING ", flags)
	ErrorLogger = log.New(os.Stdout, "ERROR ", flags)
}
