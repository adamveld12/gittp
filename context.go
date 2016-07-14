package gittp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type streamCode string

var (
	serviceRegexp          = regexp.MustCompile("(?:/info/refs\\?service=|/)(git-(?:receive|upload)-pack)$")
	errNoMatchingService   = errors.New("No matching service types found")
	errCouldNotReadReqBody = errors.New("couldn't read request body")

	packDataStreamCode = streamCode("\u0001")
	progressStreamCode = streamCode("\u0002")
	fatalStreamCode    = streamCode("\u0003")
)

type handlerContext struct {
	packetHeader
	ShouldRunHooks bool
	Advertisement  bool
	IsReceivePack  bool
	IsGetRefs      bool
	RepoExists     bool
	FullRepoPath   string
	RepoName       string
	ServiceType    string
	Input          io.Reader
	Output         io.Writer
}

// TODO needs tests
func newHandlerContext(res http.ResponseWriter, req *http.Request, repoPath string) (handlerContext, error) {
	serviceTypeStr, err := detectServiceType(req.URL)
	if err != nil {
		return handlerContext{}, err
	}

	url := req.URL
	repoName, err := parseRepoName(url.String())
	if err != nil {
		return handlerContext{}, err
	}

	advertise := req.Method == "GET"
	isGetRefs := advertise && strings.Contains(url.RequestURI(), "/info/refs?service=")
	isReceivePack := serviceTypeStr == "git-receive-pack"
	shouldRunHooks := isReceivePack && !advertise

	// the request body in a multi reader
	refsHeader, err := readPackInfo(req.Body)
	if err != nil {
		return handlerContext{}, errCouldNotReadReqBody
	}

	if len(refsHeader) <= 4 {
		shouldRunHooks = false
	}

	fullRepoPath := filepath.Join(repoPath, repoName)

	var rpr packetHeader
	if !advertise && isReceivePack {
		rpr = newPacketHeader(refsHeader)
	}

	_, ferr := os.Stat(fullRepoPath)
	fileExists := !os.IsNotExist(ferr)

	return handlerContext{
		packetHeader:   rpr,
		ServiceType:    serviceTypeStr,
		IsReceivePack:  isReceivePack,
		Advertisement:  advertise,
		ShouldRunHooks: shouldRunHooks,
		IsGetRefs:      isGetRefs,
		RepoName:       repoName,
		RepoExists:     fileExists,
		FullRepoPath:   fullRepoPath,
		Input:          io.MultiReader(bytes.NewBuffer(refsHeader), req.Body),
		Output:         io.MultiWriter(res, os.Stdout),
	}, nil
}

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

func newHookContext(ctx handlerContext) *HookContext {
	return &HookContext{
		ctx.RepoName,
		ctx.FullRepoPath,
		ctx.Branch,
		ctx.Head,
		ctx.RepoExists,
		ctx.Output,
	}
}

// HookContext represents the current context about an on going push for hook handlers. It contains the repo name, branch name, the commit hash and a sideband channel that can be used to write status update messsages to the client.
type HookContext struct {
	// Repository is the name of the repository being pushed to
	Repository   string
	FullRepoPath string
	// Branch is the name of the branch being pushed
	Branch string
	// Commit is the commit hash being pushed
	Commit string
	// RepoExists is true if the repository being pushed to exists on the remote. If this value is false and the PreReceiveHook succeeds, gittp will auto initialize a bare repo befure handling the request.
	RepoExists bool
	w          io.Writer
}

func flush(w io.Writer) {
	f, ok := w.(http.Flusher)
	if ok {
		f.Flush()
	}
}

// Fatal writes a fatal error to the git client. Useful when you want to signal that a push failed
func (h *HookContext) Fatal(msg string) error {
	payload := fmt.Sprintf("%serror: %s\n", fatalStreamCode, msg)
	_, err := h.w.Write(pktline(payload))
	return err
}

// Write writes a []byte to the git client
func (h *HookContext) Write(data []byte) (i int, e error) {
	defer flush(h.w)
	return h.w.Write(encodeWithPrefix(progressStreamCode, string(data)))
}

// Writelnf writes a string to the git client using a format string and parameters
func (h *HookContext) Writelnf(fmtString string, params ...interface{}) error {
	return h.Writeln(fmt.Sprintf(fmtString, params...))
}

// Writeln writes a string to the git client
func (h *HookContext) Writeln(text string) error {
	defer flush(h.w)
	_, err := h.w.Write(encodeWithPrefix(progressStreamCode, text+"\n"))
	return err
}
