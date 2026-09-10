package observability

import (
	"log"
	"time"
)

func Event(name string, fields map[string]any) {
	log.Printf("event=%s fields=%v", name, fields)
}

func Duration(start time.Time) float64 {
	return time.Since(start).Seconds()
}
