-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Revert the secondary indexes added in the corresponding up migration.

DROP INDEX idx_im_target_remediation_date ON IssueMatch;
DROP INDEX idx_im_created_at              ON IssueMatch;
DROP INDEX idx_im_rating                  ON IssueMatch;
DROP INDEX idx_im_acknowledged            ON IssueMatch;

DROP INDEX idx_ci_region    ON ComponentInstance;
DROP INDEX idx_ci_cluster   ON ComponentInstance;
DROP INDEX idx_ci_namespace ON ComponentInstance;
DROP INDEX idx_ci_domain    ON ComponentInstance;
DROP INDEX idx_ci_project   ON ComponentInstance;
DROP INDEX idx_ci_pod       ON ComponentInstance;
DROP INDEX idx_ci_container ON ComponentInstance;
DROP INDEX idx_ci_type      ON ComponentInstance;

DROP INDEX idx_cv_version      ON ComponentVersion;
DROP INDEX idx_cv_tag          ON ComponentVersion;
DROP INDEX idx_cv_repository   ON ComponentVersion;
DROP INDEX idx_cv_organization ON ComponentVersion;
DROP INDEX idx_cv_end_of_life  ON ComponentVersion;
DROP INDEX idx_sr_tag          ON ScannerRun;
DROP INDEX idx_sr_is_completed ON ScannerRun;
