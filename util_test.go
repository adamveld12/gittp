package gittp

import (
	"fmt"
	"net/url"
	"testing"
)

func parseURL(raw string) (parsed *url.URL) {
	parsed, _ = url.Parse(raw)
	return
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

func Test_buildPacket(t *testing.T) {
	actual := string(writePacket("# service=git-receive-pack\n"))
	expected := fmt.Sprintf("001f# service=git-receive-pack\n0000")

	if actual != expected {
			t.Errorf("expected:\n%s\nactual:\n%s\n", expected, actual)
	}
}

func Test_parseRepoName(t *testing.T) {
	testCases := map[string]string{
		"/adam/project.git/user/refs?service=git-receive-pack":      "/adam/project.git",
		"/adam/dude/project.git/user/refs?service=git-receive-pack": "/adam/dude/project.git",
		"/adam/project.git":                                         "/adam/project.git",
		"/adam/gittp.git/git-receive-pack":                          "/adam/gittp.git",
	}

	for input, expected := range testCases {
		actual, err := parseRepoName(input)

		if err != nil {
			t.Error(err)
		} else if actual != expected {
			t.Errorf("expected:\n%s\nactual:\n%s\n", expected, actual)
		}
	}

	// test err condition
	if _, err := parseRepoName("/adam/project/user/refs?service=git-receive-pack"); err == nil {
		t.Error("expected an error")
	}
}

func Test_encode(t *testing.T) {
	cases := map[string]string{
		"0010\u0002Hello world":   "Hello world",
		"0018\u0002☃woooo☃☃woooo": "☃woooo☃☃woooo",
	}

	for expected, testcase := range cases {
		actual := encode(testcase)
		if expected != actual {
			t.Errorf("expected:\n%s\nactual:\n%s\n", expected, actual)
		}
	}
}

func Test_newReceivePackResult(t *testing.T) {
	// need to get some git-receive-pack data to test with
}
