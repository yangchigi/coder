package codersdk

import "github.com/google/uuid"

type JFrogXrayScan struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Critical    int       `json:"critical"`
	High        int       `json:"high"`
	ResultsURL  string    `json:"results_url"`
}
