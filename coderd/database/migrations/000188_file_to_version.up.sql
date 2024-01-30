ALTER TABLE
	files
ADD COLUMN
	-- This is nullable, which means it is unlinked.
	template_version_id UUID REFERENCES template_versions(id) ON DELETE CASCADE;

UPDATE files SET template_version_id = template_versions.id
