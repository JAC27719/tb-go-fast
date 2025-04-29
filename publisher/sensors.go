package main

import (
	"encoding/json"
)

type ISensor interface {
	Read() ([]byte, error)
}

type BME280 struct {
	T float32
	H float32
}

func (s *BME280) Read() ([]byte, error) {
	return json.Marshal(s)
}

type CapacitiveVWC struct {
	M float32
}

func (s *CapacitiveVWC) Read() ([]byte, error) {
	return json.Marshal(s)
}
