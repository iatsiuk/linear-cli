package api

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func looksLikeUUID(s string) bool {
	return uuidRe.MatchString(s)
}

type idNode struct {
	ID string `json:"id"`
}

type nodeConnection struct {
	Nodes []idNode `json:"nodes"`
}

// ResolveTeamID resolves a team key (e.g. "ENG") or UUID to a team UUID.
// If keyOrID is already a UUID, it is returned as-is without an API call.
func ResolveTeamID(ctx context.Context, c *Client, keyOrID string) (string, error) {
	if looksLikeUUID(keyOrID) {
		return keyOrID, nil
	}

	var result struct {
		Teams nodeConnection `json:"teams"`
	}
	const q = `
		query ResolveTeam($key: String!) {
			teams(filter: { key: { eq: $key } }, first: 1) {
				nodes { id }
			}
		}`
	if err := c.Do(ctx, q, map[string]any{"key": keyOrID}, &result); err != nil {
		return "", fmt.Errorf("resolve team %q: %w", keyOrID, err)
	}
	if len(result.Teams.Nodes) == 0 {
		return "", fmt.Errorf("team %q not found", keyOrID)
	}
	return result.Teams.Nodes[0].ID, nil
}

// ResolveLabelID resolves a label name to a label UUID.
// If name is already a UUID, it is returned as-is without an API call.
// teamID is optional; when provided, the search is restricted to that team.
func ResolveLabelID(ctx context.Context, c *Client, name, teamID string) (string, error) {
	if looksLikeUUID(name) {
		return name, nil
	}

	var result struct {
		IssueLabels nodeConnection `json:"issueLabels"`
	}

	vars := map[string]any{"name": name}
	var q string
	if teamID != "" {
		q = `
			query ResolveLabel($name: String!, $teamID: ID!) {
				issueLabels(filter: { name: { eq: $name }, team: { id: { eq: $teamID } } }, first: 1) {
					nodes { id }
				}
			}`
		vars["teamID"] = teamID
	} else {
		q = `
			query ResolveLabel($name: String!) {
				issueLabels(filter: { name: { eq: $name } }, first: 1) {
					nodes { id }
				}
			}`
	}

	if err := c.Do(ctx, q, vars, &result); err != nil {
		return "", fmt.Errorf("resolve label %q: %w", name, err)
	}
	if len(result.IssueLabels.Nodes) == 0 {
		return "", fmt.Errorf("label %q not found", name)
	}
	return result.IssueLabels.Nodes[0].ID, nil
}

// ResolveUserID resolves a user name or email to a user UUID.
// If nameOrEmail is already a UUID, it is returned as-is without an API call.
func ResolveUserID(ctx context.Context, c *Client, nameOrEmail string) (string, error) {
	if looksLikeUUID(nameOrEmail) {
		return nameOrEmail, nil
	}

	var result struct {
		Users nodeConnection `json:"users"`
	}

	// try name -> displayName -> email
	const qName = `
		query ResolveUserByName($name: String!) {
			users(filter: { name: { eq: $name } }, first: 1) {
				nodes { id }
			}
		}`
	const qDisplayName = `
		query ResolveUserByDisplayName($displayName: String!) {
			users(filter: { displayName: { eq: $displayName } }, first: 1) {
				nodes { id }
			}
		}`
	const qEmail = `
		query ResolveUserByEmail($email: String!) {
			users(filter: { email: { eq: $email } }, first: 1) {
				nodes { id }
			}
		}`

	if err := c.Do(ctx, qName, map[string]any{"name": nameOrEmail}, &result); err != nil {
		return "", fmt.Errorf("resolve user %q: %w", nameOrEmail, err)
	}
	if len(result.Users.Nodes) == 0 {
		if err := c.Do(ctx, qDisplayName, map[string]any{"displayName": nameOrEmail}, &result); err != nil {
			return "", fmt.Errorf("resolve user %q: %w", nameOrEmail, err)
		}
	}
	if len(result.Users.Nodes) == 0 {
		if err := c.Do(ctx, qEmail, map[string]any{"email": nameOrEmail}, &result); err != nil {
			return "", fmt.Errorf("resolve user %q: %w", nameOrEmail, err)
		}
	}
	if len(result.Users.Nodes) == 0 {
		return "", fmt.Errorf("user %q not found", nameOrEmail)
	}
	return result.Users.Nodes[0].ID, nil
}

// ResolveStateID resolves a workflow state name to a state UUID.
// If name is already a UUID, it is returned as-is without an API call.
// teamID is optional; when provided, the search is restricted to that team.
func ResolveStateID(ctx context.Context, c *Client, name, teamID string) (string, error) {
	if looksLikeUUID(name) {
		return name, nil
	}

	var result struct {
		WorkflowStates nodeConnection `json:"workflowStates"`
	}

	vars := map[string]any{"name": name}
	var q string
	if teamID != "" {
		q = `
			query ResolveState($name: String!, $teamID: ID!) {
				workflowStates(filter: { name: { eq: $name }, team: { id: { eq: $teamID } } }, first: 1) {
					nodes { id }
				}
			}`
		vars["teamID"] = teamID
	} else {
		q = `
			query ResolveState($name: String!) {
				workflowStates(filter: { name: { eq: $name } }, first: 1) {
					nodes { id }
				}
			}`
	}

	if err := c.Do(ctx, q, vars, &result); err != nil {
		return "", fmt.Errorf("resolve workflow state %q: %w", name, err)
	}
	if len(result.WorkflowStates.Nodes) == 0 {
		return "", fmt.Errorf("workflow state %q not found", name)
	}
	return result.WorkflowStates.Nodes[0].ID, nil
}

// ResolveViewerID returns the ID of the authenticated user.
func ResolveViewerID(ctx context.Context, c *Client) (string, error) {
	var result struct {
		Viewer struct {
			ID string `json:"id"`
		} `json:"viewer"`
	}
	const q = `query { viewer { id } }`
	if err := c.Do(ctx, q, nil, &result); err != nil {
		return "", fmt.Errorf("resolve viewer: %w", err)
	}
	if result.Viewer.ID == "" {
		return "", fmt.Errorf("viewer not found")
	}
	return result.Viewer.ID, nil
}

// ResolveProjectStatusID resolves a project status type string (e.g. "started") to a status UUID.
// If typeOrID is already a UUID, it is returned as-is without an API call.
func ResolveProjectStatusID(ctx context.Context, c *Client, typeOrID string) (string, error) {
	if looksLikeUUID(typeOrID) {
		return typeOrID, nil
	}

	var result struct {
		ProjectStatuses struct {
			Nodes []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"nodes"`
		} `json:"projectStatuses"`
	}
	const q = `query { projectStatuses(first: 250) { nodes { id type } } }`
	if err := c.Do(ctx, q, nil, &result); err != nil {
		return "", fmt.Errorf("resolve project status %q: %w", typeOrID, err)
	}
	normalized := strings.ToLower(typeOrID)
	for _, s := range result.ProjectStatuses.Nodes {
		if strings.ToLower(s.Type) == normalized {
			return s.ID, nil
		}
	}
	return "", fmt.Errorf("project status %q not found", typeOrID)
}

// ResolveIssueID resolves an issue identifier (e.g. "ENG-727") or UUID to an
// issue UUID. If idOrKey is already a UUID, it is returned as-is without an
// API call. The Linear API accepts both UUID and identifier forms in the
// top-level issue(id:) field, so the input is forwarded directly.
func ResolveIssueID(ctx context.Context, c *Client, idOrKey string) (string, error) {
	if looksLikeUUID(idOrKey) {
		return idOrKey, nil
	}

	var result struct {
		Issue *idNode `json:"issue"`
	}
	const q = `
		query ResolveIssue($id: String!) {
			issue(id: $id) { id }
		}`
	if err := c.Do(ctx, q, map[string]any{"id": idOrKey}, &result); err != nil {
		return "", fmt.Errorf("resolve issue %q: %w", idOrKey, err)
	}
	if result.Issue == nil || result.Issue.ID == "" {
		return "", fmt.Errorf("issue %q not found", idOrKey)
	}
	return result.Issue.ID, nil
}

// ResolveCustomViewID resolves a custom view name or URL slug to a UUID.
// UUID input is returned as-is. Name input is resolved via exact match.
// If no name match is found, the original input is returned unchanged so the
// API can handle it as a URL slug (e.g. "my-team-bugs").
func ResolveCustomViewID(ctx context.Context, c *Client, nameOrID string) (string, error) {
	if looksLikeUUID(nameOrID) {
		return nameOrID, nil
	}

	var result struct {
		CustomViews nodeConnection `json:"customViews"`
	}
	const q = `
		query ResolveCustomView($name: String!) {
			customViews(filter: { name: { eq: $name } }, first: 1) {
				nodes { id }
			}
		}`
	if err := c.Do(ctx, q, map[string]any{"name": nameOrID}, &result); err != nil {
		return "", fmt.Errorf("resolve custom view %q: %w", nameOrID, err)
	}
	if len(result.CustomViews.Nodes) == 0 {
		// no name match; pass through so the API can resolve it as a slug
		return nameOrID, nil
	}
	return result.CustomViews.Nodes[0].ID, nil
}

// ResolveProjectID resolves a project name to a project UUID.
// If name is already a UUID, it is returned as-is without an API call.
func ResolveProjectID(ctx context.Context, c *Client, name string) (string, error) {
	if looksLikeUUID(name) {
		return name, nil
	}

	var result struct {
		Projects nodeConnection `json:"projects"`
	}
	const q = `
		query ResolveProject($name: String!) {
			projects(filter: { name: { eq: $name } }, first: 1) {
				nodes { id }
			}
		}`
	if err := c.Do(ctx, q, map[string]any{"name": name}, &result); err != nil {
		return "", fmt.Errorf("resolve project %q: %w", name, err)
	}
	if len(result.Projects.Nodes) == 0 {
		return "", fmt.Errorf("project %q not found", name)
	}
	return result.Projects.Nodes[0].ID, nil
}
