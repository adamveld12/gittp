package gittp

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// PreReceiveHook is a func called on pre receive. This is right before a git push is processed. Returning false from this handler will cancel the push to the remote, and returning true will allow the process to continue
type PreReceiveHook func(HookContext) bool

// PostReceiveHook is a func called after git-receive-pack is ran. This is a good place to fire notifications.
type PostReceiveHook func(HookContext, io.Reader)

// ServerConfig is a configuration object for NewGitServer
type ServerConfig struct {
	// Path is the file path where pushed repositories are stored
	Path string
	// Enables debug logging
	Debug bool
	// PostReceive is a post receive hook that is ran after refs have been successfully processed. Useful for running automated builds, sending notifications etc.
	PostReceive PostReceiveHook
	// PreReceive is a pre receive hook that is ran before the repo is updated. Useful for enforcing branch naming (master only pushing).
	PreReceive PreReceiveHook
	// OnCreate is a hook called when a push causes a new repository to be created
	OnCreate PostReceiveHook
}

type gitHTTPServer struct {
	ServerConfig
}

// NewGitServer initializes a new http.Handler that can serve to a git client over HTTP. An error is returned if the specified repositories path does not exist.
func NewGitServer(config ServerConfig) (http.Handler, error) {
	config.Path, _ = filepath.Abs(config.Path)

	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		if err := os.MkdirAll(config.Path, os.ModeDir|os.ModePerm); err != nil {
			return nil, errors.New("Could not create repository path")
		}
	}

	return &gitHTTPServer{
		config,
	}, nil
}

// OnlyExistingRepositories is a pre receive hook that only accepts pushes to initialized repositories
func OnlyExistingRepositories(h HookContext) bool {
	if !h.RepoExists {
		h.Writeln("The specifed repository is not initialized")
	}

	return h.RepoExists
}

// MasterOnly is a prereceive hook that only allows pushes to master
func MasterOnly(h HookContext) bool {
	if h.Branch == "refs/heads/master" {
		return true
	}

	h.Writeln("Only pushing to master is allowed.")

	return false
}

var repoRegex = regexp.MustCompile("^(?:/[\\w]+){2}.git")

// UseGithubRepoNames enforces paths like /username/projectname.git
func UseGithubRepoNames(h HookContext) bool {
	return repoRegex.MatchString(h.Repository)
}
