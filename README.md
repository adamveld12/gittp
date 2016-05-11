# GITTP

Host your own git server over HTTP.

Comes in CLI and Library flavors.

I used [this doc](https://www.kernel.org/pub/software/scm/git/docs/technical/http-protocol.html) and this handy [blog post](http://www.michaelfcollins3.me/blog/2012/05/18/implementing-a-git-http-server.html)


## How to CLI

Simply run `gittp` at your command line after installing the binary into your `$PATH`.

Available args:
`-path`: Specify a file path where pushed repositories are stored

`-port`: Specify the port that gittp should listen on

## How to Library

There is a high level API and a low level API.

In the high level API the main object that you'll use is the `GitHTTPListener`. This type implements `http.Handler` so you can easily integrate it into an existing server.

```go
package main

import (
  "net/http"
  "github.com/adamveld12/gittp"
)

func main() {
	config := gittp.ServerConfig{
    Path: "./repositories",
    PreReceive: gittp.MasterOnlyPreReceive
  }
	handle, err := gittp.NewGitServer(config)

  log.Fatal(http.ListenAndServe(":80", handle))
}
```

