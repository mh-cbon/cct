
[![travis Status](https://travis-ci.org/mh-cbon/cct.svg?branch=master)](https://travis-ci.org/mh-cbon/cct) 
[![appveyor Status](https://ci.appveyor.com/api/projects/status/github/mh-cbon/cct?branch=master&svg=true)](https://ci.appveyor.com/project/mh-cbon/cct) [![Go Report Card](https://goreportcard.com/badge/github.com/mh-cbon/cct)](https://goreportcard.com/report/github.com/mh-cbon/cct) [![GoDoc](https://godoc.org/github.com/mh-cbon/cct?status.svg)](http://godoc.org/github.com/mh-cbon/cct) [![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

# cct
# Usage

### $ cct -version

  Show version

### $ cct -help

  Show this help

### $ cct -add <bucket> <cmd>

  Add <cmd> to given <bucket>

### $ cct -wait [-verbose] [-keep] [-immediate] [-json] <bucket>

  Wait for <bucket> commands completion, prints command results.
  When a command of the bucket is finished, it is removed.

  - -immediate: prevent the program to wait for bucket completion before returning.
  - -keep: prevent the program to remove finished commands of the bucket being queried.

### $ cct -backend [-verbose] [-timeout n]

  Start the backend to execute commands concurrently.

  The backend automatically exists after duration <n> when
  - tasks list is empty
  - tasks are finished

  The backend listens to http activity to delay the timeout.

  -timeout n: duration length before the backend exits.

# Example

### add a tasks to bucket 1

    cct -add 1 ls -al
    cct -add 1 wait 10
    cct -add 1 wait 5

### wait completion of the bucket 1

    cct -wait 1

  This command will wait for the completion of all three commands added to the bucket 1

    cct -wait 1

  Running the command againg will immediately return an empty result
  as the bucket 1 has been emptied by the previous call.

### get status of the bucket 1

    cct -wait -keep -immediate -json 1

  Using __-keep__ and __-immediate__ options to query the bucket 1
  will give you the status of all tasks in this bucket.
