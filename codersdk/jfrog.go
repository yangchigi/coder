package codersdk

import "github.com/google/uuid"

type JFrogXrayScan struct {
	WorkspaceID uuid.UUID `json:"uuid"`
	Critical    int       `json:"critical"`
	High        int       `json:"high"`
	ResultsURL  string    `json:"results_url"`
}
