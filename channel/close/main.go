package main

import (
	"fmt"
	"log"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}

	c := make(chan int)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			c <- i
		}
		close(c)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range c {
			if i != 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%d", i)
		}
		fmt.Println("")
		log.Print("finished reading")
	}()

	wg.Wait()
	log.Print("completed")
}
