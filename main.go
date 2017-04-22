// Package cct is a cli program to run concurrent command lines
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// VERSION is the last build number.
var VERSION = "0.0.0"

var showLog bool

func logMsg(f string, s ...interface{}) {
	if showLog {
		fmt.Printf(f+"\n", s...)
	}
}

func main() {

	scheme := flag.String("scheme", "http", "scheme of the backend")
	host := flag.String("host", "localhost", "host of the backend")
	port := flag.String("port", "9999", "port of the backend")
	duration := flag.Int("timeout", 60, "timeout of the backend")
	backend := flag.Bool("backend", false, "run the backend")
	add := flag.Bool("add", false, "add task to the backend")
	wait := flag.Bool("wait", false, "wait task to end")
	immediate := flag.Bool("immediate", false, "do not wait for tasks to finish")
	keep := flag.Bool("keep", false, "keep tasks in the backlog")
	verbose := flag.Bool("verbose", false, "verbose")
	help := flag.Bool("help", false, "help")
	showVer := flag.Bool("version", false, "show version")
	flag.Parse()
	restargs := flag.Args()

	showLog = *verbose

	logMsg("restargs %v", restargs)
	logMsg("backend %v", *backend)

	if *help {
		showVersion()
		showHelp()
	} else if *showVer {
		showVersion()
	} else if *backend {
		runBackend(*scheme, *host, *port, restargs, *duration)
	} else if *add {
		runAdd(*scheme, *host, *port, restargs)
	} else if *wait {
		runWait(*scheme, *host, *port, restargs, *immediate, *keep)
	} else {
		showHelp()
	}
}

func showVersion() {
	fmt.Printf("cct - %v\n", VERSION)
}
func showHelp() {
	fmt.Printf(`cct is a cli program to run concurrent command lines

# Usage

#### $ cct -version

    Show version

#### $ cct -help

    Show this help

#### $ cct -add $bucket $cmd

    Add $cmd to given $bucket

#### $ cct -wait [-verbose] [-keep] [-immediate] [-json] $bucket

    Wait for <bucket> commands completion, prints command results.
    When a command of the bucket is finished and queried, it is removed.

    -immediate: prevent the program to wait for bucket completion before returning.
    -keep: prevent the program to remove finished commands of the bucket being queried.

#### $ cct -backend [-verbose] [-timeout n]

    Start the backend to execute commands concurrently.

    The backend automatically exits after duration <n> when
    - tasks list is empty
    - tasks are finished

    The backend watches for all http activities and delay the timeout.

    -timeout n: duration length before the backend exits automatically.

# Example

#### add tasks to bucket 1

    cct -add 1 ls -al
    cct -add 1 sleep 10
    cct -add 1 sleep 5

#### wait completion of the bucket 1

    cct -wait 1

  This command will wait for the completion of all three commands added to the bucket 1

    cct -wait 1

  Running the command again will return immediately,
	the response is an empty list
	as the bucket 1 was flushed by the previous call.

#### query status of the bucket 1

    cct -wait -keep -immediate -json 1

  Use -keep and -immediate options to only
	query the status of every commands in the bucket 1.

	Those options prevent the bucket to be emptied.
`)
}

// BucketCmd is a cmd to put in a bucket
type BucketCmd struct {
	Name string
	Bin  string
	Args []string
}

// Task is a cmd to run
type Task struct {
	ID     int
	Cmd    BucketCmd
	Status CmdStatus
}

// CmdStatus is the status of a command
type CmdStatus struct {
	Started bool
	Ended   bool
	Output  []byte
	Error   string
}

func runBackend(scheme, host, port string, restargs []string, duration int) {
	timedout := make(chan bool)
	activity := make(chan bool)
	getCmds := make(chan getTasksQuery)
	go httpstart(port, activity, getCmds)
	go starttimeout(timedout, activity, getCmds, duration)
	<-timedout
}

func starttimeout(timedout chan bool, activity chan bool, getCmds chan getTasksQuery, duration int) {
	for {
		select {
		case <-time.After(time.Second * time.Duration(duration)):
			logMsg("starttimeout timedout %v", true)
			//
			opt := GetTasksOptions{
				Bucket: "",
				Keep:   true,
			}
			query := getTasksQuery{opt: opt, ret: make(chan []Task)}
			getCmds <- query
			tasks := <-query.ret
			if tasksAllDone(tasks) {
				timedout <- true
				return
			}
		case <-activity:
			logMsg("starttimeout activity %v", true)
		}
	}
}

type getTasksQuery struct {
	opt GetTasksOptions
	ret chan []Task
}

func httpstart(port string, activity chan bool, getCmds chan getTasksQuery) {
	startCmds := make(chan BucketCmd)
	updateCmds := make(chan Task)
	tasks := map[int]Task{}
	go func() {
		for todo := range startCmds {
			logMsg("startCmds todo %v", todo)
			task := Task{Cmd: todo}
			task.ID = len(tasks)
			task.Status.Started = true
			go func() {
				updateCmds <- task
				cmd := exec.Command(task.Cmd.Bin, task.Cmd.Args...)
				out, err := cmd.CombinedOutput()
				if err != nil {
					task.Status.Error = err.Error()
				}
				task.Status.Output = out
				task.Status.Ended = true
				updateCmds <- task
			}()
		}
	}()
	go func() {
		for {
			select {
			case task := <-updateCmds:
				logMsg("updateCmds task %v", task)
				tasks[task.ID] = task
			case query := <-getCmds:
				logMsg("getCmds query %v", query)
				ret := []Task{}
				for _, t := range tasks {
					if t.Cmd.Name == query.opt.Bucket || query.opt.Bucket == "" {
						ret = append(ret, t)
						if query.opt.Keep == false {
							delete(tasks, t.ID)
						}
					}
				}
				logMsg("getCmds %v", ret)
				// sort here.
				query.ret <- ret
			}
		}
	}()
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		logMsg("httpstart ping %v", true)
		activity <- true
	})
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		logMsg("httpstart add %v", true)
		activity <- true

		decoder := json.NewDecoder(r.Body)
		var cmd BucketCmd
		err := decoder.Decode(&cmd)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Body.Close()
		startCmds <- cmd
	})
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		logMsg("httpstart tasks %v", true)

		activity <- true

		decoder := json.NewDecoder(r.Body)
		var opt GetTasksOptions
		err := decoder.Decode(&opt)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Body.Close()

		query := getTasksQuery{opt: opt, ret: make(chan []Task)}
		getCmds <- query
		tasks := <-query.ret
		logMsg("httpstart done %v", true)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	})
	logMsg("httpstart port %v", port)
	err := http.ListenAndServe(":"+port, nil)
	logMsg("httpstart err %v", err)
}

func runAdd(scheme, host, port string, restargs []string) {

	if len(restargs) == 0 {
		log.Fatal("missing bucket name")
	} else if len(restargs) == 1 {
		log.Fatal("missing command to execute")
	}

	pingURL := fmt.Sprintf("%v://%v:%v/ping", scheme, host, port)
	logMsg("runAdd pingURL %v", pingURL)

	if ping(pingURL) == false {
		fork(scheme, host, port)
		pingUntilReady(pingURL)
	}

	cmd := BucketCmd{
		Name: restargs[0],
		Bin:  restargs[1],
		Args: restargs[2:],
	}
	logMsg("runAdd cmd %v", cmd)

	addURL := fmt.Sprintf("%v://%v:%v/add", scheme, host, port)
	logMsg("runAdd addURL %v", addURL)

	if err := addTask(addURL, cmd); err != nil {
		log.Fatal(err)
	}
}

func fork(scheme, host, port string) {
	bin, err := os.Executable()
	if err != nil {
		panic(err)
	}
	logMsg("fork bin %v", bin)
	args := []string{"-backend", "-scheme", scheme, "-host", host, "-port", port}
	logMsg("fork args %v", args)
	cmd := exec.Command(bin, args...)
	if cmd.Start() == nil {
		logMsg("fork Start %v", true)
		cmd.Process.Release()
	}
}

func addTask(url string, cmd BucketCmd) error {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(cmd)
	_, err := http.Post(url, "application/json; charset=utf-8", b)
	return err
}

func ping(url string) bool {
	response, err := http.Get(url)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}

func pingUntilReady(url string) {
	for ping(url) == false {
	}
}

// GetTasksOptions is the options to use when getting tasks
type GetTasksOptions struct {
	Bucket    string
	Immediate bool
	Keep      bool
}

func tasksAllDone(tasks []Task) bool {
	allDone := true
	for _, t := range tasks {
		if t.Status.Ended == false {
			allDone = false
			break
		}
	}
	logMsg("allDone %v", allDone)
	return allDone
}

func runWait(scheme, host, port string, restargs []string, immediate, keep bool) {

	if len(restargs) == 0 {
		log.Fatal("missing bucket name")
	}

	tasksURL := fmt.Sprintf("%v://%v:%v/tasks", scheme, host, port)
	opt := GetTasksOptions{
		Bucket:    restargs[0],
		Immediate: immediate,
		Keep:      keep,
	}
	tasks := getTasks(tasksURL, opt)
	for {
		if immediate {
			break
		}
		if tasksAllDone(tasks) {
			break
		}
		tasks = getTasks(tasksURL, opt)
	}
	for _, task := range tasks {
		fmt.Printf("$ %v %v\n", task.Cmd.Bin, strings.Join(task.Cmd.Args, " "))
		fmt.Printf("%v", string(task.Status.Output))
		fmt.Printf("%v", task.Status.Error)
	}
}

func getTasks(url string, opt GetTasksOptions) []Task {
	var ret []Task
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(opt)
	res, err := http.Post(url, "application/json; charset=utf-8", &b)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.NewDecoder(res.Body).Decode(&ret); err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	return ret
}
