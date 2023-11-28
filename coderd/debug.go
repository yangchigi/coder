package coderd

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"tailscale.com/util/singleflight"

	"github.com/coder/coder/v2/coderd/healthcheck"
	"github.com/coder/coder/v2/coderd/httpapi"
	"github.com/coder/coder/v2/coderd/httpmw"
	"github.com/coder/coder/v2/coderd/workspaceapps"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/tailnet"
)

type DebugHealthOptions struct {
	TailnetCoordinator *atomic.Pointer[tailnet.Coordinator]
	AgentProvider      workspaceapps.AgentProvider
	HealthcheckTimeout time.Duration
	HealthcheckRefresh time.Duration
	HealthcheckFunc    func(ctx context.Context, apiKey string) *healthcheck.Report
}

type DebugHealth struct {
	tailnetCoordinator *atomic.Pointer[tailnet.Coordinator]
	agentProvider      workspaceapps.AgentProvider
	healthCheckGroup   *singleflight.Group[string, *healthcheck.Report]
	healthCheckCache   atomic.Pointer[healthcheck.Report]
	healthcheckTimeout time.Duration
	healthcheckRefresh time.Duration
	healthcheckFunc    func(ctx context.Context, apiKey string) *healthcheck.Report
}

func NewDebugHealth(ctx context.Context, opts DebugHealthOptions) *DebugHealth {
	dh := &DebugHealth{
		tailnetCoordinator: opts.TailnetCoordinator,
		agentProvider:      opts.AgentProvider,
		healthCheckGroup:   &singleflight.Group[string, *healthcheck.Report]{},
		healthcheckTimeout: opts.HealthcheckTimeout,
		healthcheckRefresh: opts.HealthcheckRefresh,
		healthcheckFunc:    opts.HealthcheckFunc,
	}

	go func() {
		ticker := time.NewTicker(opts.HealthcheckRefresh)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dh.healthcheckFunc(ctx, "???????????????????????????????")
				dh.healthCheckCache.Store(nil)
			}
		}
	}()

	return dh
}

// @Summary Debug Info Wireguard Coordinator
// @ID debug-info-wireguard-coordinator
// @Security CoderSessionToken
// @Produce text/html
// @Tags Debug
// @Success 200
// @Router /debug/coordinator [get]
func (dh *DebugHealth) debugCoordinator(rw http.ResponseWriter, r *http.Request) {
	(*dh.tailnetCoordinator.Load()).ServeHTTPDebug(rw, r)
}

// @Summary Debug Info Tailnet
// @ID debug-info-tailnet
// @Security CoderSessionToken
// @Produce text/html
// @Tags Debug
// @Success 200
// @Router /debug/tailnet [get]
func (dh *DebugHealth) debugTailnet(rw http.ResponseWriter, r *http.Request) {
	dh.agentProvider.ServeHTTPDebug(rw, r)
}

// @Summary Debug Info Deployment Health
// @ID debug-info-deployment-health
// @Security CoderSessionToken
// @Produce json
// @Tags Debug
// @Success 200 {object} healthcheck.Report
// @Router /debug/health [get]
// @Param force query boolean false "Force a healthcheck to run"
func (dh *DebugHealth) debugDeploymentHealth(rw http.ResponseWriter, r *http.Request) {
	apiKey := httpmw.APITokenFromRequest(r)
	ctx, cancel := context.WithTimeout(r.Context(), dh.healthcheckTimeout)
	defer cancel()

	// Check if the forced query parameter is set.
	forced := r.URL.Query().Get("force") == "true"

	// Get cached report if it exists and the requester did not force a refresh.
	if !forced {
		if report := dh.healthCheckCache.Load(); report != nil {
			if time.Since(report.Time) < dh.healthcheckRefresh {
				formatHealthcheck(ctx, rw, r, report)
				return
			}
		}
	}

	resChan := dh.healthCheckGroup.DoChan("", func() (*healthcheck.Report, error) {
		// Create a new context not tied to the request.
		ctx, cancel := context.WithTimeout(context.Background(), dh.healthcheckTimeout)
		defer cancel()

		report := dh.healthcheckFunc(ctx, apiKey)
		dh.healthCheckCache.Store(report)
		return report, nil
	})

	select {
	case <-ctx.Done():
		httpapi.Write(ctx, rw, http.StatusNotFound, codersdk.Response{
			Message: "Healthcheck is in progress and did not complete in time. Try again in a few seconds.",
		})
		return
	case res := <-resChan:
		formatHealthcheck(ctx, rw, r, res.Val)
		return
	}
}

func formatHealthcheck(ctx context.Context, rw http.ResponseWriter, r *http.Request, hc *healthcheck.Report) {
	format := r.URL.Query().Get("format")
	switch format {
	case "text":
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)

		_, _ = fmt.Fprintln(rw, "time:", hc.Time.Format(time.RFC3339))
		_, _ = fmt.Fprintln(rw, "healthy:", hc.Healthy)
		_, _ = fmt.Fprintln(rw, "derp:", hc.DERP.Healthy)
		_, _ = fmt.Fprintln(rw, "access_url:", hc.AccessURL.Healthy)
		_, _ = fmt.Fprintln(rw, "websocket:", hc.Websocket.Healthy)
		_, _ = fmt.Fprintln(rw, "database:", hc.Database.Healthy)

	case "", "json":
		httpapi.WriteIndent(ctx, rw, http.StatusOK, hc)

	default:
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: fmt.Sprintf("Invalid format option %q.", format),
			Detail:  "Allowed values are: \"json\", \"simple\".",
		})
	}
}

// For some reason the swagger docs need to be attached to a function.
//
// @Summary Debug Info Websocket Test
// @ID debug-info-websocket-test
// @Security CoderSessionToken
// @Produce json
// @Tags Debug
// @Success 201 {object} codersdk.Response
// @Router /debug/ws [get]
// @x-apidocgen {"skip": true}
func _debugws(http.ResponseWriter, *http.Request) {} //nolint:unused

// // RegisterDebugHealthMetrics registers debug health metrics with prometheus.
// func RegisterDebugHealthMetrics(registerer prometheus.Registerer) error {
// 	accessURLHealthyGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
// 		Namespace: "coderd",
// 		Subsystem: "health",
// 		Name:      "access_url_healthy",
// 		Help:      "Access URL Health",
// 	}, []string{})
// 	err := registerer.Register(accessURLHealthyGauge)
// 	if err != nil {
// 		return err
// 	}
// 	accessURLReachableGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
// 		Namespace: "coderd",
// 		Subsystem: "health",
// 		Name:      "access_url_reachable",
// 		Help:      "Access URL Reachable",
// 	}, []string{})
// 	err = registerer.Register(accessURLReachableGauge)
// 	if err != nil {
// 		return err
// 	}
// 	accessURLStatusCodeGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
// 		Namespace: "coderd",
// 		Subsystem: "health",
// 		Name:      "access_url_status_code",
// 		Help:      "Access URL Status Code",
// 	}, []string{})
// 	err = registerer.Register(accessURLStatusCodeGauge)
// 	if err != nil {
// 		return err
// 	}
// 	accessURLResponseLengthGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
// 		Namespace: "coderd",
// 		Subsystem: "health",
// 		Name:      "access_url_response_len",
// 		Help:      "Access URL Response Length",
// 	}, []string{})
// 	err = registerer.Register(accessURLResponseLengthGauge)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
