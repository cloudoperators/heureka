-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE ComponentInstance 
    DROP FOREIGN KEY fk_componentinstance_parent_id;

ALTER TABLE ComponentInstance 
    DROP COLUMN componentinstance_parent_id;

ALTER TABLE ComponentInstance 
    MODIFY componentinstance_type enum('Unknown', 'Project', 'Server', 'SecurityGroup', 'DnsZone', 'FloatingIp', 'RbacPolicy', 'User', 'Container') default 'Unknown' null;

ALTER TABLE ComponentInstance 
    MODIFY componentinstance_component_version_id int unsigned NOT NULL;