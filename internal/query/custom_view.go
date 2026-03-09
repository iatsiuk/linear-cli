package query

const customViewFields = `
	id
	name
	description
	shared
	modelName
`

// CustomViewListQuery fetches custom views for the user.
const CustomViewListQuery = `
query CustomViewList($first: Int) {
	customViews(first: $first) {
		nodes {` + customViewFields + `}
	}
}
`

// CustomViewShowQuery fetches a single custom view by ID.
const CustomViewShowQuery = `
query CustomViewShow($id: String!) {
	customView(id: $id) {` + customViewFields + `}
}
`
