package support

import (
	"context"
	"io"
	"sync"

	"cdr.dev/slog"

	"golang.org/x/xerrors"

	"github.com/coder/coder/v2/codersdk"
)

type Bundle struct {
	DeploymentBuildInfo Section[codersdk.BuildInfoResponse]
}

type Section[T any] struct {
	Result T
	Err    error
	Logs   io.Writer
}

type Deps struct {
	Client *codersdk.Client
	Log    *slog.Logger
}

func Run(ctx context.Context, deps Deps) (*Bundle, error) {
	if deps.Client == nil {
		return nil, xerrors.Errorf("deps.Client must not be nil")
	}
	if deps.Log == nil {
		return nil, xerrors.Errorf("deps.Log must not be nil")
	}
	var res Bundle
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		res.DeploymentBuildInfo = *runBuildInfo(ctx, deps)
	}()
	wg.Wait()

	return &res, nil
}

func runBuildInfo(ctx context.Context, deps Deps) (s *Section[codersdk.BuildInfoResponse]) {
	buildInfo, err := deps.Client.BuildInfo(ctx)
	s.Result = buildInfo
	s.Err = err
	return s
}
