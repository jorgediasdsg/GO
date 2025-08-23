package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Measurement struct {
	Min   float64
	Max   float64
	Sum   float64
	Count int64
}

func main() {
	measurements, err := os.Open("measurements.txt")
	if err != nil {
		panic(err)
	}
	defer measurements.Close()

	data := make(map[string]Measurement)

	scanner := bufio.NewScanner(measurements)
	for scanner.Scan() {
		rawData := scanner.Text()
		semicolon := strings.Index(rawData, ";")
		location := rawData[:semicolon]
		rawTemp := rawData[semicolon+1:]

		temp, _ := strconv.ParseFloat(rawTemp, 64)

		measurements, ok := data[location]

		if !ok {
			measurements = Measurement{
				Min:   temp,
				Max:   temp,
				Sum:   temp,
				Count: 1,
			}
		} else {

			measurements.Min = min(measurements.Min, temp)
			measurements.Max = max(measurements.Max, temp)
			measurements.Sum += temp
			measurements.Count++
		}

		data[location] = measurements
	}

	locations := make([]string, 0, len(data))

	for name := range data {
		locations = append(locations, name)
	}

	sort.Strings(locations)

	fmt.Printf("{")
	for _, name := range locations {
		measurements := data[name]
		avg := measurements.Sum / float64(measurements.Count)
		fmt.Printf("'%s=%.1f/%.1f/%.1f', ", name, measurements.Min, avg, measurements.Max)
	}
	fmt.Printf("}\n")
}
