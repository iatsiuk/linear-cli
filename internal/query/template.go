package query

const templateFields = `
	id
	name
	type
	description
`

// TemplateListQuery fetches all templates in the organization.
const TemplateListQuery = `
query TemplateList {
	templates {` + templateFields + `}
}
`

// TemplateShowQuery fetches a single template by ID including templateData.
const TemplateShowQuery = `
query TemplateShow($id: String!) {
	template(id: $id) {` + templateFields + `
		templateData
	}
}
`
