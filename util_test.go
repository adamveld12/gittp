package gittp

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"
)

func parseURL(raw string) (parsed *url.URL) {
	parsed, _ = url.Parse(raw)
	return
}

func Test_parseRepoName(t *testing.T) {
	testCases := map[string]string{
		"/adam/project.git/info/refs?service=git-receive-pack":      "adam/project.git",
		"/adam/dude/project.git/info/refs?service=git-receive-pack": "adam/dude/project.git",
		"/adam/project.git":                                         errNotAGitRequest.Error(),
		"/adam/gittp.git/git-receive-pack":                          "adam/gittp.git",
		"/adam/gittp/info/refs?service=git-receive-pack":            "adam/gittp",
		"/adam/dude/project/git-receive-pack":                       "adam/dude/project",
		"/adamveld12/goku.git/info/refs?service=git-receive-pack":   "adamveld12/goku.git",
	}

	for input, expected := range testCases {
		actual, err := parseRepoName(input)

		if err != nil && err.Error() != expected {
			t.Error(err)
		} else if err == nil && actual != expected {
			t.Errorf("expected:\n%s\nactual:\n%s\n", expected, actual)
		}
	}

}

func Test_pktline(t *testing.T) {
	cases := map[string]string{
		"000fHello world":   "Hello world",
		"0017☃woooo☃☃woooo": "☃woooo☃☃woooo",
	}

	for expected, testcase := range cases {
		actual := pktline(testcase)
		if !bytes.Equal([]byte(expected), actual) {
			t.Errorf("expected:\n%s\nactual:\n%s\n", expected, actual)
		}
	}
}

func Test_encode(t *testing.T) {
	cases := map[string]string{
		"0010\u0002Hello world":   "Hello world",
		"0018\u0002☃woooo☃☃woooo": "☃woooo☃☃woooo",
	}

	for expected, testcase := range cases {
		actual := encodeWithPrefix(progressStreamCode, testcase)
		if !bytes.Equal([]byte(expected), actual) {
			fmt.Println("expected:", []byte(expected))
			fmt.Println("actual:  ", actual)
			t.Errorf("expected:\n%s\nactual:\n%s\n", expected, actual)
		}
	}
}

func Test_newPacketHeader(t *testing.T) {
	// need to get some git-receive-pack data to test with
	packData := []byte("00940000000000000000000000000000000000000000 68839ad5d8bedf1147c214e4897ca6ad8afbfecc refs/heads/master\x00report-status side-band-64k agent=git/2.8.30000")

	actual := newPacketHeader(packData)
	if actual.Agent != "git/2.8.3" {
		t.Error(actual.Agent)
	}

	if actual.Branch != "refs/heads/master" {
		t.Error(actual.Branch)
	}

	if actual.Last != "0000000000000000000000000000000000000000" {
		t.Error(actual.Last)
	}

	if actual.Head != "68839ad5d8bedf1147c214e4897ca6ad8afbfecc" {
		t.Error(actual.Head)
	}
}
