package coderd

import (
	"net/http"

	"github.com/coder/coder/v2/coderd/httpapi"
	"github.com/coder/coder/v2/codersdk"
)

// @Summary Get experiments
// @ID get-experiments
// @Security CoderSessionToken
// @Produce json
// @Tags General
// @Success 200 {object} codersdk.ExperimentsResponse
// @Router /experiments [get]
func (api *API) handleExperimentsGet(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	httpapi.Write(ctx, rw, http.StatusOK, codersdk.ExperimentsResponse{
		Enabled:   api.Experiments,
		Available: codersdk.ExperimentsAll,
	})
}
