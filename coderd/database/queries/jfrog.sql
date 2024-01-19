-- name: GetJFrogXrayScanByWorkspaceID :one
SELECT
	*
FROM
	jfrog_xray
WHERE
	workspace_id = $1
LIMIT
	1;

-- name: InsertJFrogXrayScanByWorkspaceID :exec
INSERT INTO 
	jfrog_xray (
		workspace_id,
		payload
	)
VALUES 
	($1, $2);
