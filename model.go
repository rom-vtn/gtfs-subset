package main

type Stop struct {
	StopId string `csv:"stop_id"`
	Lat    string `csv:"stop_lat"`
	Lon    string `csv:"stop_lon"`
}

type StopTime struct {
	StopId string `csv:"stop_id"`
	TripId string `csv:"trip_id"`
}

type Trip struct {
	TripId    string `csv:"trip_id"`
	ServiceId string `csv:"service_id"`
}
