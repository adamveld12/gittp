package gittp

import (
	"bytes"
	"net/http"
)

// func Test_newHandlerContext(t *testing.T) {
// 	testCases := []*http.Request{
// 		createRequest("POST", "/info/refs?service=git-receive-pack"),
// 		createRequest("GET", "/info/refs?service=git-receive-pack"),
// 		createRequest("POST", "/git-receive-pack"),
// 		createRequest("GET", "/git-receive-pack"),
// 		createRequest("POST", "/git-upload-pack"),
// 		createRequest("GET", "/git-upload-pack"),
// 	}
// }

func createRequest(method, url string) (req *http.Request) {
	req, _ = http.NewRequest(method, url, &bytes.Buffer{})
	return
}
