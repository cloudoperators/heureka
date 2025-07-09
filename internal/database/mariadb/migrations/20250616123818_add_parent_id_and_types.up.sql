-- SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE ComponentInstance 
    DROP FOREIGN KEY fk_component_instance_component_version;

ALTER TABLE ComponentInstance 
    MODIFY componentinstance_type enum('Unknown', 'Project', 'Server', 'SecurityGroup','SecurityGroupRule', 'DnsZone', 'FloatingIp', 'RbacPolicy', 'User', 'Container', 'RecordSet', 'ProjectConfiguration') default 'Unknown' null;

ALTER TABLE ComponentInstance 
    MODIFY componentinstance_component_version_id int unsigned null;

ALTER TABLE ComponentInstance 
    ADD CONSTRAINT fk_component_instance_component_version
        FOREIGN KEY (componentinstance_component_version_id) REFERENCES ComponentVersion (componentversion_id)
        ON UPDATE CASCADE
        ON DELETE SET NULL;