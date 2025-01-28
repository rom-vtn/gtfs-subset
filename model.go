package main

type Stop struct {
	StopId          string `csv:"stop_id"`
	ParentStationId string `csv:"parent_station"`
	Lat             string `csv:"stop_lat"`
	Lon             string `csv:"stop_lon"`
}

type StopTime struct {
	StopId string `csv:"stop_id"`
	TripId string `csv:"trip_id"`
}

type Trip struct {
	TripId    string `csv:"trip_id"`
	RouteId   string `csv:"route_id"`
	ServiceId string `csv:"service_id"`
	ShapeId   string `csv:"shape_id"`
}

// works for calendar and calendar dates alike since we don't care about other attributes
type CalendarElement struct {
	ServiceId string `csv:"service_id"`
}

type Route struct {
	RouteId string `csv:"route_id"`
}

type ShapePoint struct {
	ShapeId string `csv:"shape_id"`
}
