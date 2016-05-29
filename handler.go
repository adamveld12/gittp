package gittp

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

func (g *gitHTTPServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if g.Debug {
		log.Println(req.Method, req.URL)
	}

	header := res.Header()

	header.Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	header.Set("Pragma", "no-cache")
	header.Set("Cache-Control", "no-cache, max-age=0, must-revalidate")

	var ctx handlerContext
	var err error
	if ctx, err = newHandlerContext(g, req); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ctx.flush(res)

	handlerContentType := "result"
	if ctx.Advertisement {
		handlerContentType = "advertisement"
	}

	header.Set("Content-Type", fmt.Sprintf("application/x-%s-%s", ctx.ServiceType, handlerContentType))

	repoExists := fileExists(ctx.FullRepoPath)
	if ctx.ShouldRunHooks {
		receivePack := newReceivePackResult(ctx.Input)
		hookCtx := newHookContext(res,
			ctx.RepoName,
			receivePack,
			repoExists)

		if g.PreReceive != nil && !g.PreReceive(hookCtx) {
			hookCtx.close()
			return
		}

		if g.PostReceive != nil {
			defer func() {
				archive, _ := gitArchive(ctx.FullRepoPath, receivePack.NewRef)
				g.PostReceive(hookCtx, archive)
			}()
		}

		if !repoExists && g.OnCreate != nil {
			archive, _ := gitArchive(ctx.FullRepoPath, receivePack.NewRef)
			defer g.OnCreate(hookCtx, archive)
		}
	}

	// might be wise to make a pre receive handler for creating repos if they don't exist
	// and remove this code
	if !repoExists && ctx.ServiceType.isReceivePack() && initRepository(ctx.FullRepoPath) != nil {

		if g.Debug {
			log.Println("creating repository")
		}

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
		res.WriteHeader(http.StatusNotModified)
		return
	}
}
