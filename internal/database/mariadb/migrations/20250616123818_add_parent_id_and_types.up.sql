-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE ComponentInstance 
    DROP FOREIGN KEY fk_component_instance_component_version,
    MODIFY componentinstance_type enum('Unknown', 'Project', 'Server', 'SecurityGroup','SecurityGroupRule', 'DnsZone', 'FloatingIp', 'RbacPolicy', 'User', 'Container', 'RecordSet') default 'Unknown' null,
    ADD COLUMN componentinstance_parent_id int unsigned null AFTER componentinstance_type,
    MODIFY componentinstance_component_version_id int unsigned null,
    ADD CONSTRAINT fk_componentinstance_parent_id
        FOREIGN KEY (componentinstance_parent_id) REFERENCES ComponentInstance (componentinstance_id)
            ON UPDATE CASCADE
            ON DELETE CASCADE;
