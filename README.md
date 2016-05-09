# GITTP

Host your own git server over HTTP.

Comes in CLI and Library flavors.


## How To Run the CLI

Simply run `gittp` at your command line after installing the binary into your `$PATH`.

Available args:
`-path`: Specify a file path where pushed repositories are stored
`-port`: Specify the port that gittp should listen on

## How to use the Library

There is a high level API and a low level API.

In the high level API the main object that you'll use is the `GitHTTPListener`. This type implements `http.Handler` so you can easily integrate it into an existing server.

```go
package main

import (
  "net/http"
  "github.com/adamveld12/gittp"
)

func main(){
	gitListener := gittp.NewGitHTTPListener()
	
	if err := http.ListenAndServe(":80", gitListener); err != nil {
	  log.Fatal(err)
	}
}
```

In the low level API you get access to the `io.ReadCloser` and `io.Writer` objects that handle the actual git repo data. You are also given access to the raw `http.Request` and `http.ResponseWriter` so you can do cooler things like authentication or custom messages to the client.


