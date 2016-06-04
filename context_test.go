package gittp

import (
	"bytes"
	"net/http"
	"testing"
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

func Test_contentType(t *testing.T) {
	cases := []struct {
		isAdvertisement bool
		serviceType     string
		expected        string
	}{
		{true, "git-receive-pack", "application/x-git-receive-pack-advertisement"},
		{false, "git-receive-pack", "application/x-git-receive-pack-result"},

		{false, "git-upload-pack", "application/x-git-upload-pack-result"},
		{true, "git-upload-pack", "application/x-git-upload-pack-advertisement"},
	}

	for _, c := range cases {
		actual := contentType(c.serviceType, c.isAdvertisement)
		if actual != c.expected {
			t.Errorf("expected %v - actual %v", c.expected, actual)
		}
	}
}

func Test_detectServiceType(t *testing.T) {
	testCases := map[string]string{
		"/info/refs?service=git-receive-pack": "git-receive-pack",
		"/info/refs?service=git-upload-pack":  "git-upload-pack",
		"/info/refs":                          errNoMatchingService.Error(),
		"adam/testrepo.git/info/refs?service=git-receive-pack":       "git-receive-pack",
		"adam/test/repo/info/refs/info/refs?service=git-upload-pack": "git-upload-pack",
		"/git-receive-pack":                                          "git-receive-pack",
		"/git-upload-pack":                                           "git-upload-pack",
		"adam/test.git/git-receive-pack":                             "git-receive-pack",
		"adam/test.git/git-upload-pack":                              "git-upload-pack",
		"git-upload-pack/git-receive-pack":                           "git-receive-pack",
		"git-receive-pack/git-upload-pack":                           "git-upload-pack",
		"adam/test/git-upload-pack/git-receive-pack":                 "git-receive-pack",
		"adam/test/git":                                              errNoMatchingService.Error(),
	}

	for rawURL, expected := range testCases {
		actual, err := detectServiceType(parseURL(rawURL))

		if err != nil && err.Error() != expected {
			t.Logf("testing %s -> %s\n", rawURL, expected)
			t.Errorf("expected:\n%s\nactual:\n%s\n", expected, err.Error())
		} else if err == nil && actual != expected {
			t.Logf("testing %s -> %s\n", rawURL, expected)
			t.Errorf("expected:\n%s\nactual:\n%s\n", expected, actual)
		}
	}
}
