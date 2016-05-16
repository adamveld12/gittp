# GITTP

Host your own git server over HTTP.

Comes in CLI and Library flavors.

I used [this doc](https://www.kernel.org/pub/software/scm/git/docs/technical/http-protocol.html) and this handy [blog post](http://www.michaelfcollins3.me/blog/2012/05/18/implementing-a-git-http-server.html)


## How to CLI

Simply run `gittp` at your command line after installing the binary into your `$PATH`.

Available args:

`-path`: Specify a file path where pushed repositories are stored. This folder must exist, and gittp will throw an error if it can't be found

`-port`: Specify the port that gittp should listen on

`-masterOnly`: Only permits pushing to the master branch

`-debug`: turns on debug logging (currently WIP)

## How to Library

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
    PreReceive: gittp.MasterOnly,
    PostReceive: func(h gittp.HookContext){
      h.Writef("Woohoo! Push to %s succeeded!\n", h.Repository)
    }
  }

	handle, err := gittp.NewGitServer(config)

  log.Fatal(http.ListenAndServe(":80", handle))
}
```
