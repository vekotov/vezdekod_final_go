package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		println("Usage: ./vezdekod_go_1 filename.txt")
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

	for i, line := range lines {
		duration, err := time.ParseDuration(line)

		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Started task #" + strconv.Itoa(i) + " execution (" + line + ")")
		time.Sleep(duration)
		log.Println("Finished task #" + strconv.Itoa(i) + " execution (" + line + ")")
	}
}
