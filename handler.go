package gittp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func (g *gitHTTPServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if g.Debug {
		fmt.Println("")
		debugPrint(res, req)
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
			receivePack.Branch,
			receivePack.NewRef,
			fileExists(ctx.FullRepoPath))

		defer hookCtx.close()

		if g.PreReceive != nil && !g.PreReceive(hookCtx) {
			return
		}

		if g.PostReceive != nil {
			defer g.PostReceive(hookCtx)
		}

	}

	if err := handleMissingRepo(ctx.ServiceType, ctx.FullRepoPath); err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	output := ctx.Output
	if g.Debug {
		output = io.MultiWriter(os.Stdout, output)
	}

	if ctx.IsGetRefs {
		output.Write(writePacket(fmt.Sprintf("# service=%s\n", ctx.ServiceType)))
	}

	if err := runCmd(ctx.ServiceType,
		ctx.FullRepoPath,
		bytes.NewBuffer(ctx.Input),
		output,
		ctx.Advertisement); err != nil {
		res.WriteHeader(http.StatusNotModified)
		return
	}
}

func debugPrint(res http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.String())
}
