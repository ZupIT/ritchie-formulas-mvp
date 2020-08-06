package main

import (
	"itau/formula/pkg/formula"
	"os"
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
