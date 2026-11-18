-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

-- Add secondary indexes for frequently filtered and cursor-sorted columns.
--
-- Only FK join columns (auto-indexed by InnoDB) and UNIQUE-fronted columns are
-- indexed today, so filter/sort-only columns fall back to full table scans.
--
-- Soft-delete shape: every list and aggregation query against the entity tables
-- below always leads with `<table>_deleted_at IS NULL` (injected centrally by the
-- state filter for all of these entities). An index on the bare filter/sort
-- column therefore cannot be used on its own -- the optimizer needs deleted_at to
-- be the leading key column. So each index is a (deleted_at, column) composite:
-- deleted_at satisfies the universal equality predicate, and the second column
-- serves the filter or cursor sort that follows. ScannerRun is the exception --
-- its queries do not apply the deleted_at filter, so its indexes stay single
-- column.
--
-- InnoDB adds a secondary index online by default (built in place, reads and
-- writes continue during the build), so no explicit ALGORITHM/LOCK clause is
-- needed. We intentionally omit them: forcing ALGORITHM=INPLACE turns an
-- otherwise-online build into a hard error on tables the engine would rather
-- build another way, without changing the resulting index.

-- Tier 1: IssueMatch (largest table; filter + cursor-sort hot path)
CREATE INDEX idx_im_target_remediation_date ON IssueMatch (issuematch_deleted_at, issuematch_target_remediation_date);
CREATE INDEX idx_im_created_at              ON IssueMatch (issuematch_deleted_at, issuematch_created_at);
CREATE INDEX idx_im_rating                  ON IssueMatch (issuematch_deleted_at, issuematch_rating);
CREATE INDEX idx_im_acknowledged            ON IssueMatch (issuematch_deleted_at, issuematch_acknowledged);

-- Tier 2: ComponentInstance label columns (filtered + cursor-sorted)
--
-- The label columns are VARCHAR(1024). At utf8mb4 (4 bytes/char) a full-width
-- key column is 4096 bytes on its own -- already past InnoDB's 3072-byte max
-- key length -- so we index a 255-character prefix of each. 255 chars retains
-- more than enough selectivity for these labels while keeping the composite
-- (deleted_at + prefix) comfortably within the limit. componentinstance_type is
-- a narrow ENUM and needs no prefix.
CREATE INDEX idx_ci_region    ON ComponentInstance (componentinstance_deleted_at, componentinstance_region(255));
CREATE INDEX idx_ci_cluster   ON ComponentInstance (componentinstance_deleted_at, componentinstance_cluster(255));
CREATE INDEX idx_ci_namespace ON ComponentInstance (componentinstance_deleted_at, componentinstance_namespace(255));
CREATE INDEX idx_ci_domain    ON ComponentInstance (componentinstance_deleted_at, componentinstance_domain(255));
CREATE INDEX idx_ci_project   ON ComponentInstance (componentinstance_deleted_at, componentinstance_project(255));
CREATE INDEX idx_ci_pod       ON ComponentInstance (componentinstance_deleted_at, componentinstance_pod(255));
CREATE INDEX idx_ci_container ON ComponentInstance (componentinstance_deleted_at, componentinstance_container(255));
CREATE INDEX idx_ci_type      ON ComponentInstance (componentinstance_deleted_at, componentinstance_type);

-- Tier 3: ComponentVersion filter-only columns
CREATE INDEX idx_cv_version      ON ComponentVersion (componentversion_deleted_at, componentversion_version);
CREATE INDEX idx_cv_tag          ON ComponentVersion (componentversion_deleted_at, componentversion_tag);
CREATE INDEX idx_cv_repository   ON ComponentVersion (componentversion_deleted_at, componentversion_repository);
CREATE INDEX idx_cv_organization ON ComponentVersion (componentversion_deleted_at, componentversion_organization);
CREATE INDEX idx_cv_end_of_life  ON ComponentVersion (componentversion_deleted_at, componentversion_end_of_life);

-- ScannerRun: queries filter on these columns directly without a deleted_at
-- predicate, so single-column indexes are the right shape here.
CREATE INDEX idx_sr_tag          ON ScannerRun (scannerrun_tag);
CREATE INDEX idx_sr_is_completed ON ScannerRun (scannerrun_is_completed);
