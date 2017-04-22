# cct

[![travis Status](https://travis-ci.org/mh-cbon/cct.svg?branch=master)](https://travis-ci.org/mh-cbon/cct) 
[![appveyor Status](https://ci.appveyor.com/api/projects/status/github/mh-cbon/cct?branch=master&svg=true)](https://ci.appveyor.com/project/mh-cbon/cct) [![Go Report Card](https://goreportcard.com/badge/github.com/mh-cbon/cct)](https://goreportcard.com/report/github.com/mh-cbon/cct) [![GoDoc](https://godoc.org/github.com/mh-cbon/cct?status.svg)](http://godoc.org/github.com/mh-cbon/cct) [![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Package cct is a cli program to run concurrent command lines


# TOC
- [Install](#install)
  - [go](#go)
  - [linux rpm/deb repository](#linux-rpmdeb-repository)
  - [linux rpm/deb standalone package](#linux-rpmdeb-standalone-package)
- [Cli](#cli)
  - [cct](#cct-1)
- [Usage](#usage)
  - [$ cct -version](#-cct--version)
  - [$ cct -help](#-cct--help)
  - [$ cct -add|-a [-timeout n] $bucket $cmd](#-cct--add|-a-[-timeout-n]-bucket-cmd)
  - [$ cct -wait|-w [-verbose] [-keep|-k] [-immediate|-i] [-json|-j] $bucket](#-cct--wait|-w-[-verbose]-[-keep|-k]-[-immediate|-i]-[-json|-j]-bucket)
  - [$ cct -backend [-verbose] [-timeout n]](#-cct--backend-[-verbose]-[-timeout-n])
- [Example](#example)
  - [add tasks to bucket 1](#add-tasks-to-bucket-1)
  - [wait completion of the bucket 1](#wait-completion-of-the-bucket-1)
  - [query status of the bucket 1](#query-status-of-the-bucket-1)
  - [Release the project](#release-the-project)
- [History](#history)

# Install

Check the [release page](https://github.com/mh-cbon/cct/releases)!

#### go
```sh
go get github.com/mh-cbon/cct
```

#### linux rpm/deb repository
```sh
wget -O - https://raw.githubusercontent.com/mh-cbon/latest/master/source.sh \
| GH=mh-cbon/cct sh -xe
# or
curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/source.sh \
| GH=mh-cbon/cct sh -xe
```

#### linux rpm/deb standalone package
```sh
curl -L https://raw.githubusercontent.com/mh-cbon/latest/master/install.sh \
| GH=mh-cbon/cct sh -xe
# or
wget -q -O - --no-check-certificate \
https://raw.githubusercontent.com/mh-cbon/latest/master/install.sh \
| GH=mh-cbon/cct sh -xe
```

# Cli

## cct
cct is a cli program to run concurrent command lines

# Usage

#### $ cct -version

    Show version

#### $ cct -help

    Show this help

#### $ cct -add|-a [-timeout n] $bucket $cmd

    Add $cmd to given $bucket

#### $ cct -wait|-w [-verbose] [-keep|-k] [-immediate|-i] [-json|-j] $bucket

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

    cct -add -timeout 2 1 sleep 2
    cct -a 1 sleep 10
    cct -a 1 sleep 5

Note: using a one liner syntax would result in a totally different result!

    cct -a 1 sleep 10;cct -a 1 sleep 10

#### wait completion of the bucket 1

    cct -wait 1

  This command will wait for the completion of all three commands added to the bucket 1

    cct -w 1

  Running the command again will return immediately,
	the response is an empty list
	as the bucket 1 was flushed by the previous call.

#### query status of the bucket 1

    cct -w -k -i -j 1

  Use [-keep|-k] and [-immediate|-i] options to only
  query the status of every commands in the bucket 1.

  Those options prevent the bucket to be emptied.

  Use [-json|-j] option to return a JSON obect of []Task.

#### Release the project

```sh
gump patch -d # check
gump patch # bump
```

# History

[CHANGELOG](CHANGELOG.md)
