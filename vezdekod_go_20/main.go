package main

import (
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
		println("Usage: ./vezdekod_go_2 filename.txt")
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

	for i, line := range lines {
		wg.Add(1)
		duration, err := time.ParseDuration(line)

		if err != nil {
			log.Fatalln(err)
		}

		go func(i int, line string, duration time.Duration) {
			defer wg.Done()
			log.Println("Started task #" + strconv.Itoa(i) + " execution (" + line + ")")
			time.Sleep(duration)
			log.Println("Finished task #" + strconv.Itoa(i) + " execution (" + line + ")")
		}(i, line, duration)
	}

	wg.Wait()
}
