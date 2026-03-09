package query

// OrganizationQuery fetches the current user's organization.
const OrganizationQuery = `
query Organization {
	organization {
		id
		name
		urlKey
		logoUrl
	}
}
`
