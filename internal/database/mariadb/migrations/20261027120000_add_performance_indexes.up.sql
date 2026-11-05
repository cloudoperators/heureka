-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- POC: performance indexes on soft-delete / join columns.
-- Every list/count query filters `<prefix>_deleted_at IS NULL`, and the heavy MV
-- procedures + per-service relation queries join through IssueMatch. None of these
-- columns were indexed, forcing full scans on the large tables.

-- IssueMatch (8.9M rows): counts/lists filter deleted_at; covering (deleted_at, id).
CREATE INDEX idx_issuematch_deleted_id
    ON IssueMatch (issuematch_deleted_at, issuematch_id);

-- IssueMatch: per-service relation joins filter deleted_at and join on
-- component_instance_id (no status filter) — component_instance_id must follow
-- deleted_at directly so the join can seek.
CREATE INDEX idx_issuematch_deleted_ci_issue
    ON IssueMatch (issuematch_deleted_at, issuematch_component_instance_id, issuematch_issue_id);

-- IssueMatch: MV aggregations filter status='new' AND deleted_at IS NULL, join on
-- component_instance_id / issue_id, and read rating. All columns are in the key so
-- the procedures run index-only (measured 5.1s -> 1.3s). status must be in the key
-- for covering even though it is non-selective.
CREATE INDEX idx_issuematch_mv
    ON IssueMatch (issuematch_deleted_at, issuematch_status, issuematch_component_instance_id, issuematch_issue_id, issuematch_rating);

-- Issue (275K): deleted_at filter on counts/lists.
CREATE INDEX idx_issue_deleted_id
    ON Issue (issue_deleted_at, issue_id);

-- IssueVariant (258K): deleted_at filter + join on issue_id.
CREATE INDEX idx_issuevariant_deleted_issue
    ON IssueVariant (issuevariant_deleted_at, issuevariant_issue_id);

-- ComponentInstance: deleted_at filter + join on service_id (per-service relations).
CREATE INDEX idx_componentinstance_deleted_service
    ON ComponentInstance (componentinstance_deleted_at, componentinstance_service_id);
