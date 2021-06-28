package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("specify TimeZone")
		os.Exit(1)
	}

	timeZone := os.Args[1]

	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(loc)
}
