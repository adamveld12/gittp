package gittp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// PreReceiveHook is a func called on pre receive. This is right before a git push is processed. Returning false from this handler will cancel the push to the remote, and returning true will allow the process to continue
type PreReceiveHook func(HookContext) bool

// PostReceiveHook is a func called after git-receive-pack is ran. This is a good place to fire notifications.
type PostReceiveHook func(HookContext, io.Reader)

// PreCreateHook is a func called before a missing repository is created. Returning false from this handler will prevent a new repository from being created.
type PreCreateHook func(string) bool

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

	// PreCreate is a hook called when a push causes a new repository to be created. This hook is ran before the repo is created.
	PreCreate PreCreateHook
}

// NewGitServer initializes a new http.Handler that can serve to a git client over HTTP. An error is returned if the specified repositories path does not exist.
func NewGitServer(config ServerConfig) (http.Handler, error) {
	config.Path, _ = filepath.Abs(config.Path)

	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		if err := os.MkdirAll(config.Path, os.ModeDir|os.ModePerm); err != nil {
			return nil, errors.New("Could not create repository path")
		}
	}

	if config.PreCreate == nil {
		config.PreCreate = DenyCreateRepo
	}

	if config.PreReceive == nil {
		config.PreReceive = NoopPreReceive
	}

	return &gitHTTPServer{
		config,
	}, nil
}

type gitHTTPServer struct{ ServerConfig }

func (g *gitHTTPServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if g.Debug {
		log.Println(req.Method, req.URL)
	}

	header := res.Header()

	header.Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	header.Set("Pragma", "no-cache")
	header.Set("Cache-Control", "no-cache, max-age=0, must-revalidate")

	ctx, err := newHandlerContext(req, g.Path)
	if err != nil {
		if g.Debug {
			log.Println("could not create handler context", err)
		}
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ctx.flush(res)

	header.Set("Content-Type", contentType(ctx.ServiceType, ctx.Advertisement))

	repoExists := fileExists(ctx.FullRepoPath)

	if ctx.ShouldRunHooks {
		ok, hookContinuation := g.runHooks(ctx, res, repoExists)
		defer hookContinuation()
		if !ok {
			return
		}
	}

	if err := g.createRepoIfMissing(res, ctx, repoExists); err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if ctx.IsGetRefs {
		ctx.Output.Write(writePacket(fmt.Sprintf("# service=%s\n", ctx.ServiceType)))
	}

	if err := runCmd(ctx.ServiceType,
		ctx.FullRepoPath,
		bytes.NewBuffer(ctx.Input),
		ctx.Output,
		ctx.Advertisement); err != nil {

		if g.Debug {
			log.Println("an error occurred running", ctx.ServiceType, err)
		}
		res.WriteHeader(http.StatusNotModified)
		return
	}
}

func (g *gitHTTPServer) runHooks(ctx handlerContext, res http.ResponseWriter, repoExists bool) (bool, func()) {
	receivePack := newReceivePackResult(ctx.Input)
	hookCtx := newHookContext(res, ctx, repoExists)

	flush := func() {}

	if !g.PreReceive(hookCtx) {
		if g.Debug {
			log.Println("pre receive hook failed")
		}
		return false, flush
	}

	if g.PostReceive != nil {
		return true, func() {
			defer flush()
			archive, _ := gitArchive(ctx.FullRepoPath, receivePack.NewRef)
			g.PostReceive(hookCtx, archive)
		}
	}

	return true, flush
}

func (g *gitHTTPServer) createRepoIfMissing(res http.ResponseWriter, ctx handlerContext, repoExists bool) error {
	shouldRunCreate := !repoExists && ctx.IsReceivePack

	if !shouldRunCreate {
		return nil
	}

	if g.PreCreate(ctx.RepoName) {
		if err := initRepository(ctx.FullRepoPath); err != nil {
			log.Println("Could not initialize repository", err)
			return err
		} else if g.Debug {
			log.Println("creating repository")
		}
	} else {
		if g.Debug {
			log.Print("pushing is disallowed")
		}
		return errors.New("Cannot create repository")
	}

	return nil
}
