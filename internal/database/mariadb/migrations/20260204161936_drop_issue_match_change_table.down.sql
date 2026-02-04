--  SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
--  SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS IssueMatchChange (
    issuematchchange_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    issuematchchange_activity_id INT UNSIGNED NOT NULL,
    issuematchchange_issue_match_id INT UNSIGNED NOT NULL,
    issuematchchange_action ENUM('add', 'remove') NOT NULL,
    issuematchchange_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP() NOT NULL,
    issuematchchange_created_by INT UNSIGNED NULL,
    issuematchchange_deleted_at TIMESTAMP NULL,
    issuematchchange_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP() NOT NULL ON UPDATE CURRENT_TIMESTAMP(),
    issuematchchange_updated_by INT UNSIGNED NULL,
    CONSTRAINT fk_issuematchchange_activity FOREIGN KEY (issuematchchange_activity_id) REFERENCES Activity (activity_id),
    CONSTRAINT fk_issuematchchange_issue_match FOREIGN KEY (
        issuematchchange_issue_match_id
    ) REFERENCES IssueMatch (issuematch_id),
    CONSTRAINT fk_issuematchchange_created_by FOREIGN KEY (issuematchchange_created_by) REFERENCES User (user_id),
    CONSTRAINT fk_issuematchchange_updated_by FOREIGN KEY (issuematchchange_updated_by) REFERENCES User (user_id)
);