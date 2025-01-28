package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"strconv"
)

// processAllObjects parses all rows of the given filename in the input feed, then tests them against the filter function. if filter function passes, all row fields will be written to the file in the output feed (if an output feed has been given). The passing function will also be run with the element that passes the filter.
func processAllObjects[T any](inputFeed *zip.ReadCloser, outputFeed *zip.Writer, feedFilename string, filter func(T) bool, passing func(T), required bool) error {
	//DEBUG STUFFIES
	if outputFeed != nil {
		fmt.Printf("processing feedFilename with rewrite: %v\n", feedFilename)
	} else {
		fmt.Printf("processing feedFilename without rewrite: %v\n", feedFilename)
	}

	//check beforehand to not create an empty file on the output
	if !feedHasFile(inputFeed, feedFilename) {
		if required {
			return fmt.Errorf("missing required file: %s", feedFilename)
		}
		return nil //allowed otherwise
	}

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
		err = nil
	}
	if err != nil {
		return err
	}
	if reader.csvWriter != nil {
		reader.csvWriter.Flush()
		return reader.csvWriter.Error() //weird csv writer interface "get error after call" thingy
	}
	return nil
}

func rawCopy(inputFeed *zip.ReadCloser, outputFeed *zip.Writer, feedFilename string, required bool) error {
	fmt.Printf("raw copying feedFilename: %v\n", feedFilename)
	for _, file := range inputFeed.File {
		if file.Name == feedFilename {
			return outputFeed.Copy(file)
		}
	}
	//if file not found
	if required {
		return fmt.Errorf("missing required file in source feed: %s", feedFilename)
	}
	return nil
}

func processFeed(params Params) error {
	inputFeed, err := zip.OpenReader(params.inputFilename)
	if err != nil {
		return err
	}

	outputFile, err := os.OpenFile(params.outputFilename, os.O_WRONLY|os.O_CREATE, os.ModeAppend)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	outputFeed := zip.NewWriter(outputFile)
	defer outputFeed.Close()

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

	err = processAllObjects(inputFeed, nil, "stops.txt", isStopInBox, addStopIdToBox, true)
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
	err = processAllObjects(inputFeed, nil, "stop_times.txt", stopTimeAtStopInBox, addTripIdToMap, true)
	if err != nil {
		return err
	}

	//add trips going through box
	tripIdGoesThroughBox := func(trip Trip) bool {
		_, ok := tripIdsInBox[trip.TripId]
		return ok
	}
	usedServiceIds := make(map[string]struct{})
	usedRouteIds := make(map[string]struct{})
	usedShapeIds := make(map[string]struct{})
	addIdsInTrip := func(trip Trip) {
		usedServiceIds[trip.ServiceId] = struct{}{}
		usedRouteIds[trip.RouteId] = struct{}{}
		usedShapeIds[trip.ShapeId] = struct{}{}
	}
	err = processAllObjects(inputFeed, outputFeed, "trips.txt", tripIdGoesThroughBox, addIdsInTrip, true)
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
	err = processAllObjects(inputFeed, outputFeed, "stop_times.txt", stopTimeHasCorrectId, addStopIdsToPassedThrough, true)
	if err != nil {
		return err
	}

	//go once through stops to add parent station ids
	stopIsPassedThrough := func(s Stop) bool {
		_, ok := stopIdsPassedThrough[s.StopId]
		return ok
	}
	finalIncludedStopIds := make(map[string]struct{})
	addToFinalIncluded := func(s Stop) {
		finalIncludedStopIds[s.StopId] = struct{}{}
		finalIncludedStopIds[s.ParentStationId] = struct{}{}
	}
	err = processAllObjects(inputFeed, nil, "stops.txt", stopIsPassedThrough, addToFinalIncluded, true)
	if err != nil {
		return err
	}
	//then actually add all those stops
	isFinalIncluded := func(s Stop) bool {
		_, ok := finalIncludedStopIds[s.StopId]
		return ok
	}
	err = processAllObjects(inputFeed, outputFeed, "stops.txt", isFinalIncluded, func(Stop) {}, true)
	if err != nil {
		return err
	}

	//do calendar and calendar dates
	hasUsedServiceId := func(ce CalendarElement) bool {
		_, ok := usedServiceIds[ce.ServiceId]
		return ok
	}
	err = processAllObjects(inputFeed, outputFeed, "calendar.txt", hasUsedServiceId, func(CalendarElement) {}, false)
	if err != nil {
		return err
	}
	err = processAllObjects(inputFeed, outputFeed, "calendar_dates.txt", hasUsedServiceId, func(CalendarElement) {}, false)
	if err != nil {
		return err
	}

	//process routes
	hasUsedRouteId := func(r Route) bool {
		_, ok := usedRouteIds[r.RouteId]
		return ok
	}
	err = processAllObjects(inputFeed, outputFeed, "routes.txt", hasUsedRouteId, func(Route) {}, true)
	if err != nil {
		return err
	}

	//process shapes
	hasUsedShapeId := func(s ShapePoint) bool {
		_, ok := usedShapeIds[s.ShapeId]
		return ok
	}
	err = processAllObjects(inputFeed, outputFeed, "shapes.txt", hasUsedShapeId, func(ShapePoint) {}, false)
	if err != nil {
		return err
	}

	var rawCopyFiles = map[string]bool{
		"agency.txt":       true,
		"feed_info.txt":    false, //fixme uhm actually it's "conditionally required" 🤓
		"attributions.txt": false,
		"levels.txt":       false,
	}

	for fileName, required := range rawCopyFiles {
		err := rawCopy(inputFeed, outputFeed, fileName, required)
		if err != nil {
			return err
		}
	}

	return nil
}
