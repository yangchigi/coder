-- name: GetJFrogXrayScanByWorkspaceAndAgentID :one
SELECT
	*
FROM
	jfrog_xray
WHERE
	agent_id = $1
AND
	workspace_id = $2
LIMIT
	1;

-- name: InsertJFrogXrayScanByWorkspaceAndAgentID :exec
INSERT INTO 
	jfrog_xray (
		agent_id,
		workspace_id,
		payload
	)
VALUES 
	($1, $2, $3);
