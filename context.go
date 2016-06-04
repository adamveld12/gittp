package gittp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	serviceRegexp          = regexp.MustCompile("(?:/info/refs\\?service=|/)(git-(?:receive|upload)-pack)$")
	errNoMatchingService   = errors.New("No matching service types found")
	errCouldNotReadReqBody = errors.New("couldn't read request body")
)

func contentType(serviceType string, isAdvertisement bool) string {
	handlerContentType := "result"

	if isAdvertisement {
		handlerContentType = "advertisement"
	}

	return fmt.Sprintf("application/x-%s-%s", serviceType, handlerContentType)
}

func detectServiceType(url *url.URL) (string, error) {
	match := serviceRegexp.FindStringSubmatch(url.RequestURI())
	if len(match) < 2 {
		return "", errNoMatchingService
	}

	return match[1], nil
}

type handlerContext struct {
	ShouldRunHooks bool
	Advertisement  bool
	IsReceivePack  bool
	IsGetRefs      bool
	FullRepoPath   string
	RepoName       string
	ServiceType    string
	Input          []byte
	Output         io.Writer
}

func (h handlerContext) flush(res http.ResponseWriter) error {
	if _, err := io.Copy(res, h.Output.(io.Reader)); err != nil {
		return err
	}

	return nil
}

func newHandlerContext(req *http.Request, repoPath string) (handlerContext, error) {
	serviceTypeStr, err := detectServiceType(req.URL)
	if err != nil {
		return handlerContext{}, err
	}

	reqPath := req.URL.String()

	repoName, err := parseRepoName(reqPath)
	if err != nil {
		return handlerContext{}, err
	}

	advertise := req.Method == "GET"
	isGetRefs := advertise && strings.Contains(req.URL.RequestURI(), "/info/refs?service=")
	isReceivePack := serviceTypeStr == "git-receive-pack"
	shouldRunHooks := isReceivePack && !advertise

	reqDataBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return handlerContext{}, errCouldNotReadReqBody
	}

	if bytes.Equal(reqDataBytes, []byte("0000")) {
		shouldRunHooks = false
	}

	return handlerContext{
		ServiceType:    serviceTypeStr,
		IsReceivePack:  isReceivePack,
		Advertisement:  advertise,
		ShouldRunHooks: shouldRunHooks,
		IsGetRefs:      isGetRefs,
		RepoName:       repoName,
		FullRepoPath:   filepath.Join(repoPath, repoName),
		Input:          reqDataBytes,
		Output:         &bytes.Buffer{},
	}, nil
}

// HookContext represents the current context about an on going push for hook handlers. It contains the repo name, branch name, the commit hash and a sideband channel that can be used to write status update messsages to the client.
type HookContext struct {
	flusher http.Flusher
	writer  io.Writer
	// Repository is the name of the repository being pushed to
	Repository string
	// Branch is the name of the branch being pushed
	Branch string
	// Commit is the commit hash being pushed
	Commit string
	// RepoExists is true if the repository being pushed to exists on the remote. If this value is false and the PreReceiveHook succeeds, gittp will auto initialize a bare repo befure handling the request.
	RepoExists bool
	// Authorization is the authorization header's value used in the request.
	Authorization string
}

func (h HookContext) close() {
	h.writer.Write([]byte("0000"))
	h.flusher.Flush()
}

// Write writes a []byte to the git client
func (h HookContext) Write(data []byte) (i int, e error) {
	defer h.flusher.Flush()

	return h.writer.Write(encodeBytes(defaultStreamCode, data))
}

// Writef writes a string to the git client using a format string and parameters
func (h HookContext) Writef(fmtString string, params ...interface{}) error {
	_, err := h.Write([]byte(fmt.Sprintf(fmtString, params...)))
	return err
}

// Writeln writes a string to the git client
func (h HookContext) Writeln(text string) error {
	_, err := h.Write([]byte(text + "\n"))
	return err
}
