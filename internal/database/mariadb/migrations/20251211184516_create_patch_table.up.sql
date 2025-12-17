-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE Patch (
    patch_id                INT UNSIGNED NOT NULL AUTO_INCREMENT,

    patch_service_id        INT UNSIGNED NOT NULL,
    patch_component_version_id INT UNSIGNED NOT NULL,

	patch_scan_id           INT UNSIGNED NOT NULL,
	patch_scan_uuid         UUID NOT NULL UNIQUE,
	patch_scan_tag          VARCHAR(255) NOT NULL,
	patch_scan_start_run    TIMESTAMP NOT NULL,
	patch_scan_end_run      TIMESTAMP NOT NULL,

    patch_created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    patch_created_by        INT UNSIGNED NULL,
    patch_updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    patch_updated_by        INT UNSIGNED NULL,
    patch_deleted_at        TIMESTAMP NULL DEFAULT NULL,

    PRIMARY KEY (patch_id),

    CONSTRAINT fk_patch_service
        FOREIGN KEY (patch_service_id) REFERENCES Service(service_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_patch_component_version
        FOREIGN KEY (patch_component_version_id) REFERENCES ComponentVersion(componentversion_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_patch_service ON Patch (patch_service_id);
CREATE INDEX idx_patch_component_version ON Patch (patch_component_version_id);
