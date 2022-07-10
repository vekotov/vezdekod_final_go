package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

const REQ_NUM = 20000
const DURATION_PER_REQ = 5 * time.Second
const GOROUTINES_NUMBER = 8
const MAX_DIFF = 10 * time.Millisecond

type GetTimeResponse struct {
	TimeLeft string `json:"time_left"`
}

type ScheduleTaskElem struct {
	Id       int64  `json:"id"`
	Duration string `json:"duration"`
}

type GetScheduleResponse struct {
	TaskList []ScheduleTaskElem `json:"task_list"`
}

func main() {
	startTime := time.Now().UnixNano()

	message := map[string]interface{}{
		"duration": DURATION_PER_REQ.String(),
		"async":    false,
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = http.Post("http://localhost:8080/add", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	endTime := time.Now().UnixNano()
	diffTime := time.Duration(endTime - startTime)
	println("/add sync actual time spent: " + diffTime.String())
	println("/add sync theoretical time spent: " + DURATION_PER_REQ.String())

	if diffTime > DURATION_PER_REQ-MAX_DIFF && diffTime < DURATION_PER_REQ+MAX_DIFF {
		println("/add sync tests passed!")
	} else {
		log.Fatalln("/add sync test failed - difference more than " + MAX_DIFF.String())
	}

	var wg sync.WaitGroup
	channelMaxGoroutines := make(chan bool, GOROUTINES_NUMBER)

	for i := 0; i < GOROUTINES_NUMBER; i++ {
		channelMaxGoroutines <- true
	}

	for i := 0; i < REQ_NUM; i++ {
		wg.Add(1)
		go func() {
			<-channelMaxGoroutines

			message := map[string]interface{}{
				"duration": DURATION_PER_REQ.String(),
				"async":    true,
			}

			bytesRepresentation, err := json.Marshal(message)
			if err != nil {
				log.Fatalln(err)
			}

			resp, err := http.Post("http://localhost:8080/add", "application/json", bytes.NewBuffer(bytesRepresentation))
			if err != nil {
				log.Fatalln(err)
			}

			if resp.StatusCode != http.StatusOK {
				log.Fatalln("Test failed - async /add returned non-200 code")
			}

			wg.Done()
			channelMaxGoroutines <- true
		}()
	}

	wg.Wait()
	println("test async /add passed!")

	resp, err := http.Get("http://localhost:8080/time")
	if err != nil {
		log.Fatalln(err)
	}

	var response GetTimeResponse

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)

	if err != nil {
		log.Fatalln(err)
	}

	println("actual time: " + response.TimeLeft)
	println("theoretical time: " + (REQ_NUM * DURATION_PER_REQ).String())

	dur, _ := time.ParseDuration(response.TimeLeft)
	theor := REQ_NUM * DURATION_PER_REQ
	if dur > theor-MAX_DIFF && dur < theor+MAX_DIFF {
		println("/time test passed!")
	} else {
		log.Fatalln("/time test failed - difference more than " + MAX_DIFF.String())
	}

	// todo: schedule ручка тест

	resp, err = http.Get("http://localhost:8080/schedule")
	if err != nil {
		log.Fatalln(err)
	}

	var scheduleResp GetScheduleResponse

	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&scheduleResp)
	if err != nil {
		log.Fatalln(err)
	}

	if len(scheduleResp.TaskList) == REQ_NUM {
		println("/schedule size test passed!")
	} else {
		log.Fatalln("/schedule size test not passed!")
	}

	for i, elem := range scheduleResp.TaskList {
		if int64(i) != elem.Id-2 {
			log.Fatalln("/schedule id test not passed!")
		}

		if elem.Duration != DURATION_PER_REQ.String() {
			log.Fatalln("/schedule duration test not passed!")
		}
	}

	println("/schedule tests passed!")

}
