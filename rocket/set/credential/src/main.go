package main

import (
	"hello/pkg/hello"
	"os"
)

func main() {
	loadInputs().Run()
}

func loadInputs() hello.Inputs {
	return hello.Inputs{
		Username: os.Getenv("USERNAME"),
		Password: os.Getenv("PASSWORD"),
		Provider: os.Getenv("PROVIDER"),
	}
}
