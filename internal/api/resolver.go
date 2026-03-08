package api

import (
	"context"
	"fmt"
	"regexp"
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

	// try matching name first; if that returns nothing, try email
	const qName = `
		query ResolveUserByName($name: String!) {
			users(filter: { name: { eq: $name } }, first: 1) {
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
		// retry by email
		result.Users.Nodes = nil
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
