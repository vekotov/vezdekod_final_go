package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		println("Usage: ./vezdekod_go_3 filename.txt")
		return
	}
	fileName := os.Args[1]

	b, err := ioutil.ReadFile(fileName)
	// can file be opened?
	if err != nil {
		log.Fatalln(err)
	}

	str := string(b)

	lines := strings.Split(str, "\n\n")

	var wg sync.WaitGroup

	print("enter processor number: ")
	var processorNum int
	scanned, err := fmt.Scanf("%d", &processorNum)
	if err != nil {
		log.Fatalln(err)
	}
	if scanned < 1 {
		println("Error: no processor number specified")
	}
	if processorNum < 1 {
		println("Processor number must be greater than 0")
	}

	syncChannel := make(chan bool, processorNum)
	for i := 0; i < processorNum; i++ {
		syncChannel <- true
	}

	for i, line := range lines {
		wg.Add(1)
		duration, err := time.ParseDuration(line)

		if err != nil {
			log.Fatalln(err)
		}

		go func(i int, line string, duration time.Duration) {
			<-syncChannel
			defer wg.Done()
			log.Println("Started task #" + strconv.Itoa(i) + " execution (" + line + ")")
			time.Sleep(duration)
			log.Println("Finished task #" + strconv.Itoa(i) + " execution (" + line + ")")
			syncChannel <- true
		}(i, line, duration)
	}

	wg.Wait()
}
