package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const (
	authToken = "NXXJgLZd8wCp-FgqvKXmxQL6pf7OkY-FstHaJ1XPtexJnFEqC1v8rjwSAPC9uLtTuMu0OyvkDz6aXF6fFfnAPw=="
	bucket    = "bnn"
	org       = "Bnn, Inc."
)

type (
	Series struct {
		Measurement string            `json:"measurement"`
		Field       string            `json:"field"`
		Tags        map[string]string `json:"tags"`
	}

	cmdHandler func(args []string) int
)

func printUsage() {
	fmt.Printf(`influx-example <subcommand> <subcommand args...>`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	var handleCmd cmdHandler

	switch os.Args[1] {
	case "write":
		handleCmd = handleWriteCommand
	default:
		printUsage()
		os.Exit(1)
	}

	os.Exit(handleCmd(os.Args[2:]))
}

func printWriteCommandUsage() {
	fmt.Printf("influx-example write [-from <RFC3339 Timestamp>] [-to <RFC3339 Timestamp>] [-duration <Duration>] <Series Definition File>\n")
}
func handleWriteCommand(args []string) int {
	var (
		from, to string
		duration time.Duration
	)

	f := flag.NewFlagSet("write", flag.ExitOnError)
	f.StringVar(&from, "from", "", "Time from which data point starts")
	f.StringVar(&to, "to", "", "Time to which data point ends")
	f.DurationVar(&duration, "duration", time.Second, "Data Point duration")

	if err := f.Parse(args); err != nil {
		printWriteCommandUsage()
		return 1
	}

	if f.NArg() < 1 {
		printWriteCommandUsage()
		return 1
	}

	var s Series
	definitionFile := f.Arg(0)
	bs, err := ioutil.ReadFile(definitionFile)
	if err != nil {
		fmt.Printf("cannot read %s properly: %s\n", definitionFile, err)
		return 1
	}
	if err = json.Unmarshal(bs, &s); err != nil {
		fmt.Printf("failed to load Series definition: %s\n", err)
		return 1
	}

	var begin time.Time
	if from != "" {
		begin, err = time.Parse(time.RFC3339, from)
		if err != nil {
			fmt.Printf("invalid <from> timestamp format: %s", err)
			return 1
		}
	} else {
		begin = time.Now().Add(-time.Minute)
	}

	var end time.Time
	if to != "" {
		end, err = time.Parse(time.RFC3339, to)
		if err != nil {
			fmt.Printf("invalid <to> timestamp formt: %s", err)
		}
	} else {
		end = time.Now()
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	client := influxdb2.NewClient("http://localhost:8086", authToken)
	writeAPI := client.WriteAPI(org, bucket)

	fmt.Printf("start writing data points to %v\n", s)
	for t := begin; t.Before(end); t = t.Add(duration) {
		v := r.Int63()
		p := influxdb2.NewPoint(s.Measurement, s.Tags, map[string]interface{}{
			s.Field: v,
		}, t)
		writeAPI.WritePoint(p)
		writeAPI.Flush()
		fmt.Printf("%s: %d\n", t.Format(time.RFC3339), v)
	}

	return 0
}
