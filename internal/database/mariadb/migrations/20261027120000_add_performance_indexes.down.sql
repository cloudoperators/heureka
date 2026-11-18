-- SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX idx_issuematch_deleted_id ON IssueMatch;
DROP INDEX idx_issuematch_deleted_ci_issue ON IssueMatch;
DROP INDEX idx_issuematch_mv ON IssueMatch;
DROP INDEX idx_issue_deleted_id ON Issue;
DROP INDEX idx_issuevariant_deleted_issue ON IssueVariant;
DROP INDEX idx_componentinstance_deleted_service ON ComponentInstance;
