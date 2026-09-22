ALTER TABLE maintenance_templates
ADD COLUMN make VARCHAR(50) NOT NULL DEFAULT '',
ADD COLUMN model VARCHAR(50) NOT NULL DEFAULT '',
ADD COLUMN variant VARCHAR(50),
ADD COLUMN year_start INT,
ADD COLUMN year_end INT;

CREATE INDEX idx_maintenance_templates_make ON maintenance_templates(make);
CREATE INDEX idx_maintenance_templates_model ON maintenance_templates(model);
