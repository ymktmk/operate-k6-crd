package main

import (
	"log"
	"os"

	"github.com/ymktmk/apply-k6-crd/pkg/k6"
)

func main() {

	template := os.Getenv("INPUT_TEMPLATE")
	if len(template) == 0 {
		fail("the INPUT_TEMPLATE has not been set")
	}

	vus := os.Getenv("INPUT_VUS")
	duration := os.Getenv("INPUT_DURATION")
	rps := os.Getenv("INPUT_RPS")
	parallelism := os.Getenv("INPUT_PARALLELISM")
	file := os.Getenv("INPUT_FILE")

	k, err := k6.NewK6(template, vus, duration, rps, parallelism, file)
	if err != nil {
		fail(err.Error())
	}

	err = k.CreateK6()
	if err != nil {
		fail(err.Error())
	}

}

func fail(err string) {
	log.Printf("Error: %s\n", err)
	os.Exit(1)
}
