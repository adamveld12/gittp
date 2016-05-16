package gittp

import (
	"bytes"
	"fmt"
	"net/http"
)

func (g *gitHTTPServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if g.Debug {
		fmt.Println("")
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

	if ctx.ShouldRunHooks {
		receivePack := newReceivePackResult(ctx.Input)
		hookCtx := newHookContext(res,
			ctx.RepoName,
			receivePack,
			fileExists(ctx.FullRepoPath))

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
	}

	if err := handleMissingRepo(ctx.ServiceType, ctx.FullRepoPath); err != nil {
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
