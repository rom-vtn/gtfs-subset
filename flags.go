package main

import (
	"flag"
	"fmt"
)

func parseFlags() (Params, error) {
	params := Params{}
	const UNSET_COORD = 500 //invalid
	const UNSET_STRING = "" //invalid
	flag.Float64Var(&params.minLat, "minlat", UNSET_COORD, "Min latitude")
	flag.Float64Var(&params.maxLat, "maxlat", UNSET_COORD, "Max latitude")
	flag.Float64Var(&params.minLon, "minlon", UNSET_COORD, "Min longitude")
	flag.Float64Var(&params.maxLon, "maxlon", UNSET_COORD, "Max longitude")
	flag.StringVar(&params.inputFilename, "input", UNSET_STRING, "Input feed ZIP file")
	flag.StringVar(&params.outputFilename, "output", UNSET_STRING, "Output feed ZIP file")
	flag.Parse()

	isValidLat := func(lat float64) error {
		if -90 <= lat && lat <= 90 {
			return nil
		}
		return fmt.Errorf("invalid latitude: %f", lat)
	}
	isValidLon := func(lon float64) error {
		if -180 <= lon && lon <= 180 {
			return nil
		}
		return fmt.Errorf("invalid longitude: %f", lon)
	}
	isValidString := func(fileName string) error {
		if fileName == "" {
			return fmt.Errorf("invalid filename: %s", fileName)
		}
		return nil
	}

	err := isValidLat(params.minLat)
	if err != nil {
		return Params{}, err
	}
	err = isValidLat(params.maxLat)
	if err != nil {
		return Params{}, err
	}
	err = isValidLon(params.minLon)
	if err != nil {
		return Params{}, err
	}
	err = isValidLon(params.maxLon)
	if err != nil {
		return Params{}, err
	}
	err = isValidString(params.inputFilename)
	if err != nil {
		return Params{}, err
	}
	err = isValidString(params.outputFilename)
	if err != nil {
		return Params{}, err
	}

	return params, nil
}
