-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS Comment (
    comment_id            INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    comment_text          LONGTEXT NOT NULL,
    comment_issuematch_id INT UNSIGNED NOT NULL,
    comment_created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    comment_created_by    INT UNSIGNED NULL,
    comment_deleted_at    TIMESTAMP NULL,
    comment_updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    comment_updated_by    INT UNSIGNED NULL,
    CONSTRAINT fk_comment_issuematch FOREIGN KEY (comment_issuematch_id) REFERENCES IssueMatch(issuematch_id) ON DELETE CASCADE,
    CONSTRAINT fk_comment_author     FOREIGN KEY (comment_created_by)    REFERENCES User(user_id)
);
