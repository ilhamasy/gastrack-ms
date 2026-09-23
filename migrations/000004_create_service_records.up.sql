CREATE TABLE service_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    service_date TIMESTAMP NOT NULL,
    odometer_km INTEGER NOT NULL,
    workshop_name VARCHAR(255),
    total_cost DECIMAL(12, 2) DEFAULT 0.0,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE service_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_record_id UUID NOT NULL REFERENCES service_records(id) ON DELETE CASCADE,
    maintenance_id UUID REFERENCES vehicle_maintenance(id) ON DELETE SET NULL,
    item_name VARCHAR(255) NOT NULL,
    brand VARCHAR(255),
    product VARCHAR(255),
    part_number VARCHAR(255),
    quantity DECIMAL(8, 2) NOT NULL DEFAULT 1.0,
    cost DECIMAL(12, 2) NOT NULL DEFAULT 0.0,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
