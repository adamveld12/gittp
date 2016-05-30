package gittp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	errCouldNotReadReqBody = errors.New("couldn't read request body")
)

type handlerContext struct {
	*gitHTTPServer
	ShouldRunHooks bool
	Advertisement  bool
	FullRepoPath   string
	RepoName       string
	ServiceType    serviceType
	Input          []byte
	IsGetRefs      bool
	Output         io.Writer
}

func (h handlerContext) flush(res http.ResponseWriter) error {
	if _, err := io.Copy(res, h.Output.(io.Reader)); err != nil {
		return err
	}

	return nil
}

func newHandlerContext(sv *gitHTTPServer, req *http.Request) (handlerContext, error) {
	serviceTypeStr, err := detectServiceType(req.URL)
	if err != nil {
		return handlerContext{}, errors.New("")
	}

	reqPath := req.URL.String()

	repoName, err := parseRepoName(reqPath)
	if err != nil {
		return handlerContext{}, err
	}

	reqDataBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return handlerContext{}, errCouldNotReadReqBody
	}

	advertise := req.Method == "GET"
	isGetRefs := advertise && strings.Contains(req.URL.RequestURI(), "/info/refs?service=")
	serviceType := serviceType(serviceTypeStr)
	shouldRunHooks := serviceType.isReceivePack() && !advertise

	if bytes.Equal(reqDataBytes, []byte("0000")) {
		shouldRunHooks = false
	}

	return handlerContext{
		gitHTTPServer:  sv,
		ServiceType:    serviceType,
		Advertisement:  advertise,
		ShouldRunHooks: shouldRunHooks,
		IsGetRefs:      isGetRefs,
		RepoName:       repoName,
		FullRepoPath:   filepath.Join(sv.Path, repoName),
		Input:          reqDataBytes,
		Output:         &bytes.Buffer{},
	}, nil
}

type serviceType string

func (s serviceType) isReceivePack() bool {
	return s == "git-receive-pack"
}

func (s serviceType) isUploadPack() bool {
	return s == "git-upload-pack"
}

func (s serviceType) String() string {
	return string(s)
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
	// Authorization is the authorization header used in the request.
	Authorization string
}

func (h HookContext) close() {
	h.writer.Write([]byte("0000"))
	h.flusher.Flush()
}

// Writes a []byte to the sideband
func (h HookContext) Write(data []byte) (i int, e error) {
	defer h.flusher.Flush()
	return h.writer.Write(encodeBytes(defaultStreamCode, data))
}

// Writef writes a string to the SideChannel using a format string and parameters
func (h HookContext) Writef(fmtString string, params ...interface{}) error {
	return h.Writeln(fmt.Sprintf(fmtString, params...))
}

// Write writes a string to the SideChannel
func (h HookContext) Writeln(text string) error {
	_, err := h.Write([]byte(text + "\n"))
	return err
}
