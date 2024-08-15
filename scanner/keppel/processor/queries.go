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
	CreateComponentVersionMatchQuery = `
	`
)
