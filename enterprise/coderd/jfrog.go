package coderd

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/httpapi"
	"github.com/coder/coder/v2/codersdk"
)

func (api *API) postJFrogXrayScan(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
	)

	var req codersdk.JFrogXrayScan
	if !httpapi.Read(ctx, rw, r, &req) {
		return
	}

	payload, err := json.Marshal(req)
	if err != nil {
		httpapi.InternalServerError(rw, err)
		return
	}

	err = api.Database.InsertJFrogXrayScanByWorkspaceID(ctx, database.InsertJFrogXrayScanByWorkspaceIDParams{
		WorkspaceID: req.WorkspaceID,
		Payload:     payload,
	})
	if err != nil {
		httpapi.InternalServerError(rw, err)
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, codersdk.Response{
		Message: "Successfully inserted XRay scan!",
	})
}

func (api *API) jFrogXrayScan(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		wid = r.URL.Query().Get("workspace_id")
	)

	id, err := uuid.Parse(wid)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "'workspace_id' must be a valid UUID.",
		})
		return
	}

	scan, err := api.Database.GetJFrogXrayScanByWorkspaceID(ctx, id)
	if xerrors.Is(err, sql.ErrNoRows) {
		httpapi.RouteNotFound(rw)
		return
	}
	if err != nil {
		httpapi.InternalServerError(rw, err)
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, scan.Payload)
}

func (api *API) jfrogEnabledMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		api.entitlementsMu.RLock()
		enabled := api.entitlements.Features[codersdk.FeatureMultipleExternalAuth].Enabled
		api.entitlementsMu.RUnlock()

		if !enabled {
			httpapi.RouteNotFound(rw)
			return
		}

		next.ServeHTTP(rw, r)
	})
}
