package processor

// GraphQL queries
const (
	GetIssueIdQuery = `
		query ($filter: IssueFilter) {
			Issues (
				filter: $filter,
			) {
				totalCount
				edges {
					node {
						id
					}
				}
			}
		}
	`
	CreateIssueQuery = `
		mutation ($input: IssueInput!) {
			createIssue (
				input: $input
			) {
				id
				primaryName
				description
				type
			}
		}
	`
	CreateIssueVariantQuery = `
		mutation ($input: IssueVariantInput!) {
			createIssueVariant (
				input: $input
			) {
				id
				secondaryName
				issueId
			}
		}
	`

	GetIssueRepositoryIdQuery = `
		query ($filter: IssueRepositoryFilter) {
			IssueRepositories (
				filter: $filter,
			) {
				totalCount
				edges {
					node {
						id
					}
				}
			}
		}
	`

	CreateIssueRepositoryQuery = `
		mutation ($input: IssueRepositoryInput!) {
			createIssueRepository (
				input: $input
			) {
				id
				name
				url
			}
		}

	`
)
