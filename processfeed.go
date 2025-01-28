package main

import (
	"archive/zip"
	"io"
	"os"
	"strconv"
)

// processAllObjects parses all rows of the given filename in the input feed, then tests them against the filter function. if filter function passes, all row fields will be written to the file in the output feed (if an output feed has been given). The passing function will also be run with the element that passes the filter.
func processAllObjects[T any](inputFeed *zip.ReadCloser, outputFeed *zip.Writer, feedFilename string, filter func(T) bool, passing func(T)) error {
	//add writer if an output feed is given
	var writer io.Writer
	if outputFeed != nil {
		w, err := outputFeed.Create(feedFilename)
		if err != nil {
			return err
		}
		writer = w
	}

	reader, err := newObjectReader[T](inputFeed, feedFilename, writer)
	if err != nil {
		return err
	}

	for {
		var current T
		err := reader.read(&current)
		if err != nil {
			break
		}
		if !filter(current) {
			continue
		}
		//always run passing function
		passing(current)

		//also add to output file if it exists
		if reader.csvWriter != nil {
			//write last record
			reader.csvWriter.Write(reader.decoder.Record())
		}
	}
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func processFeed(params Params) error {
	inputFeed, err := zip.OpenReader(params.inputFilename)
	if err != nil {
		return err
	}

	outputFile, err := os.OpenFile(params.outputFilename, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	outputFeed := zip.NewWriter(outputFile)

	//add stops in box
	isStopInBox := func(stop Stop) bool {
		lat, err := strconv.ParseFloat(stop.Lat, 64)
		if err != nil {
			return false
		}
		lon, err := strconv.ParseFloat(stop.Lon, 64)
		if err != nil {
			return false
		}
		if !(params.minLat <= lat && lat <= params.maxLat) {
			return false
		}
		if !(params.minLon <= lon && lon <= params.maxLon) {
			return false
		}
		return true
	}
	stopIdsInBox := make(map[string]struct{})
	addStopIdToBox := func(s Stop) {
		stopIdsInBox[s.StopId] = struct{}{}
	}

	err = processAllObjects(inputFeed, nil, "stops.txt", isStopInBox, addStopIdToBox)
	if err != nil {
		return err
	}

	//add trip IDs going through stops in box
	tripIdsInBox := make(map[string]struct{})
	stopTimeAtStopInBox := func(st StopTime) bool {
		_, ok := stopIdsInBox[st.StopId]
		return ok
	}
	addTripIdToMap := func(st StopTime) {
		tripIdsInBox[st.TripId] = struct{}{}
	}
	err = processAllObjects(inputFeed, nil, "stop_times.txt", stopTimeAtStopInBox, addTripIdToMap)
	if err != nil {
		return err
	}

	//add trips going through box
	tripIdGoesThroughBox := func(trip Trip) bool {
		_, ok := tripIdsInBox[trip.TripId]
		return ok
	}
	usedServiceIds := make(map[string]struct{})
	addIdsInTrip := func(trip Trip) {
		usedServiceIds[trip.ServiceId] = struct{}{}
	}
	err = processAllObjects(inputFeed, outputFeed, "trips.txt", tripIdGoesThroughBox, addIdsInTrip)
	if err != nil {
		return err
	}

	//trips OK, now add stop times
	stopTimeHasCorrectId := func(st StopTime) bool {
		_, ok := tripIdsInBox[st.TripId]
		return ok
	}
	stopIdsPassedThrough := make(map[string]struct{})
	addStopIdsToPassedThrough := func(st StopTime) {
		stopIdsPassedThrough[st.StopId] = struct{}{}
	}
	err = processAllObjects(inputFeed, outputFeed, "stop_times.txt", stopTimeHasCorrectId, addStopIdsToPassedThrough)
	if err != nil {
		return err
	}

	//now add stops
	stopIsPassedThrough := func(s Stop) bool {
		_, ok := stopIdsPassedThrough[s.StopId]
		return ok
	}
	err = processAllObjects(inputFeed, outputFeed, "stops.txt", stopIsPassedThrough, func(Stop) {})
	if err != nil {
		return err
	}

	//TODO calendar

	return nil
}
