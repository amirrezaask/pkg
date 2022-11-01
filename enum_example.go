package main

//go:generate gox $GOFILE

// gox:enum Started Arrived Finished
type RideState int
