package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Task struct {
	Id         int64         `json:"id"`
	Duration   time.Duration `json:"duration"`
	ended      bool
	returnChan chan bool
}

type SumChangerTask struct {
	add      bool
	duration time.Duration
}

type AddTaskRequest struct {
	Duration string `json:"duration"`
	Async    bool   `json:"async"`
}

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

var lastUpdatedDuration int64
var taskCreationChan chan Task
var sumDuration time.Duration
var sumDurationLock sync.RWMutex
var jobListLock sync.RWMutex
var jobList []ScheduleTaskElem

func main() {
	sumDuration = 0

	sumModifierChan := make(chan SumChangerTask, 1000)
	taskCreationChan = make(chan Task, 1000)

	workerChan := make(chan Task, 10000000)

	jobListManagerChan := make(chan Task, 1000)
	jobList = make([]ScheduleTaskElem, 0)

	go func() {
		for {
			task := <-jobListManagerChan
			jobListLock.Lock()
			if task.ended {
				jobList = jobList[1:]
			} else {
				jobList = append(jobList, ScheduleTaskElem{Id: task.Id, Duration: task.Duration.String()})
			}
			jobListLock.Unlock()
		}
	}()

	var lastTaskId int64
	lastTaskId = 0

	// оркестратор
	go func() {
		for {
			task := <-taskCreationChan
			if !task.ended {
				task.Id = lastTaskId + 1
				lastTaskId++
			}
			jobListManagerChan <- task
			if task.ended {
				sumModifierChan <- SumChangerTask{add: false, duration: task.Duration}
				if task.returnChan != nil {
					task.returnChan <- true
				}
			} else {
				sumModifierChan <- SumChangerTask{add: true, duration: task.Duration}
				workerChan <- task
			}
		}
	}()

	// изменятель суммы
	go func() {
		for {
			durMod := <-sumModifierChan
			if durMod.add {
				sumDurationLock.Lock()
				sumDuration = sumDuration + durMod.duration
				lastUpdatedDuration = time.Now().UnixNano()
				sumDurationLock.Unlock()
			} else {
				sumDurationLock.Lock()
				sumDuration = sumDuration - durMod.duration
				lastUpdatedDuration = time.Now().UnixNano()
				sumDurationLock.Unlock()
			}
		}
	}()

	// воркер
	go func() {
		for {
			task := <-workerChan
			log.Println("starting task #" + strconv.FormatInt(task.Id, 10) + " (" + task.Duration.String() + ")")
			time.Sleep(task.Duration)
			task.ended = true
			taskCreationChan <- task
			log.Println("finished task #" + strconv.FormatInt(task.Id, 10) + " (" + task.Duration.String() + ")")
		}
	}()

	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/time", getTimeHandler)
	http.HandleFunc("/schedule", getScheduleHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req AddTaskRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	dur, err := time.ParseDuration(req.Duration)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	task := Task{Duration: dur, ended: false}
	if !req.Async {
		channel := make(chan bool)
		task.returnChan = channel
		taskCreationChan <- task
		<-channel
	} else {
		taskCreationChan <- task
	}
	w.WriteHeader(http.StatusOK)
}

func getTimeHandler(w http.ResponseWriter, r *http.Request) {
	// time left is max(0, durationSum - (currTime - lastUpdate))
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sumDurationLock.Lock()
	durationSum := sumDuration
	lastUpdate := lastUpdatedDuration
	sumDurationLock.Unlock()

	left := durationSum - time.Duration(time.Now().UnixNano()-lastUpdate)
	if left < 0 {
		left = 0
	}

	resp := GetTimeResponse{TimeLeft: left.String()}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func getScheduleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jobListLock.Lock()
	json.NewEncoder(w).Encode(GetScheduleResponse{TaskList: jobList})
	jobListLock.Unlock()
}
