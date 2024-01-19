CREATE TABLE jfrog_xray (
    workspace_id uuid PRIMARY KEY REFERENCES workspaces(id) ON DELETE CASCADE,
    payload jsonb NOT NULL DEFAULT '{}'
);
