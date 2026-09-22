DROP INDEX IF EXISTS idx_maintenance_templates_model;
DROP INDEX IF EXISTS idx_maintenance_templates_make;

ALTER TABLE maintenance_templates
DROP COLUMN IF EXISTS year_end,
DROP COLUMN IF EXISTS year_start,
DROP COLUMN IF EXISTS variant,
DROP COLUMN IF EXISTS model,
DROP COLUMN IF EXISTS make;
