# GITTP

[![Go Report Card](https://goreportcard.com/badge/github.com/adamveld12/gittp)](https://goreportcard.com/report/github.com/adamveld12/gittp)
[![GoDoc](https://godoc.org/github.com/adamveld12/gittp?status.svg)](http://godoc.org/github.com/adamveld12/gittp)
[![Build Status](https://semaphoreci.com/api/v1/adamveld12/gittp/branches/master/badge.svg)](https://semaphoreci.com/adamveld12/gittp)

Host your own git server over HTTP. Effortlessly hook into pre and post receive events and write status updates back to the client.

Comes in CLI and Library flavors.

I used [this doc](https://www.kernel.org/pub/software/scm/git/docs/technical/http-protocol.html) and this handy [blog post](http://www.michaelfcollins3.me/blog/2012/05/18/implementing-a-git-http-server.html)


## How to CLI

Simply run `gittp` at your command line after installing the binary into your `$PATH`.

Available args:

`-port`: The port that gittp listens on

`-path`: Specify a file path where pushed repositories are stored. If this folder doesn't exist, gittp will create it for you

`-masterOnly`: Only permit pushing to the master branch

`-autocreate`: Auto create repositories if they have not been created

`-debug`: turns on debug logging

## How to Library

Install:
```
go get gopkg.in/adamveld12/gittp.v1
```

This lib follows http.Handler conventions. I purposely do not include any authentication, since there are many http basic authentication modules out there to use.


```go
package main

import (
  "net/http"
  "github.com/adamveld12/gittp"
)

func main() {
	config := gittp.ServerConfig{
    Path: "./repositories",
    PreCreate: gittp.UseGithubRepoNames,
    PreReceive: gittp.MasterOnly,
    PostReceive: func(h gittp.HookContext, archive io.Reader){
      h.Writef("Woohoo! Push to %s succeeded!\n", h.Branch)
    }
  }

  handle, _ := gittp.NewGitServer(config)
  log.Fatal(http.ListenAndServe(":80", handle))
}
```
