package main

import (
	"log"
)

type Params struct {
	minLat, maxLat, minLon, maxLon float64
	inputFilename, outputFilename  string
}

func main() {
	params, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	err = processFeed(params)
	if err != nil {
		log.Fatal(err)
	}
}
