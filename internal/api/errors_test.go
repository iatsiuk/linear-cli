package api

import "testing"

func TestGraphQLError_Code(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		extensions map[string]any
		want       string
	}{
		{"nil extensions", nil, ""},
		{"with code", map[string]any{"code": "RATELIMITED"}, "RATELIMITED"},
		{"code not string", map[string]any{"code": 42}, ""},
		{"other extensions", map[string]any{"other": "value"}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e := GraphQLError{Extensions: tc.extensions}
			if got := e.Code(); got != tc.want {
				t.Errorf("Code() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGraphQLErrors_Error(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		errors GraphQLErrors
		want   string
	}{
		{"single", GraphQLErrors{{Message: "not found"}}, "not found"},
		{"multiple", GraphQLErrors{{Message: "error one"}, {Message: "error two"}}, "error one; error two"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.errors.Error(); got != tc.want {
				t.Errorf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}
