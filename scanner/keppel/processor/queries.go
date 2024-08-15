package processor

const (
	CreateComponentQuery = `
		mutation ($input: ComponentInput!) {
			createComponent (
				input: $input
			) {
				id
				name
				type
			}
		}
	`
	CreateComponentVersionQuery = `
		mutation ($input: ComponentVersionInput!) {
			createComponentVersion (
				input: $input
			) {
				id
				version
				componentId
			}
		}
	`
	AddComponentVersionToIssueQuery = `
		mutation ($issueId: ID!, $componentVersionId: ID!) {
			addComponentVersionToIssue (
				issueId: $issueId,
				componentVersionId: $componentVersionId
			) {
				id
			}
		}
	`
	ListComponentsQuery = `
		query ($filter: ComponentFilter, $first: Int) {
			Components (
				filter: $filter,
				first: $first,
			) {
				edges {
					node {
						id
						name
						type
					}
				}
			}
		}
	`
	ListComponentVersionsQuery = `
		query ($filter: ComponentVersionFilter, $first: Int) {
			ComponentVersions (
				filter: $filter,
				first: $first,
			) {
				edges {
					node {
						id
						version
					}
					cursor
				}
			}
		}
	`
	ListIssueQuery = `
		query ($filter: IssueFilter, $first: Int) {
			Issues (
				filter: $filter,
				first: $first,
			) {
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
)
