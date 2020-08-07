package main

import (
	"os"
	"rocket/formula/pkg/formula"
)

func main() {
	loadInputs().Run()
}

func loadInputs() formula.Inputs {
	return formula.Inputs{
		Username: os.Getenv("USERNAME"),
		Password: os.Getenv("PASSWORD"),
	}
}
