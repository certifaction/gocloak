package gocloak_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/certifaction/gocloak/v14"
)

func TestGetOrganizationMembers_SendsBriefRepresentation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		params    gocloak.GetMembersParams
		wantQuery url.Values
	}{
		"unset": {
			params:    gocloak.GetMembersParams{},
			wantQuery: url.Values{},
		},
		"false": {
			params:    gocloak.GetMembersParams{BriefRepresentation: gocloak.BoolP(false)},
			wantQuery: url.Values{"briefRepresentation": {"false"}},
		},
		"true": {
			params:    gocloak.GetMembersParams{BriefRepresentation: gocloak.BoolP(true)},
			wantQuery: url.Values{"briefRepresentation": {"true"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var gotPath string
			var gotQuery url.Values
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotQuery = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[{"id":"u1","requiredActions":["UPDATE_PASSWORD"],"membershipType":"UNMANAGED"}]`))
			}))
			defer server.Close()

			client := gocloak.NewClient(server.URL)
			members, err := client.GetOrganizationMembers(context.Background(), "token", "realm", "org-1", tc.params)
			require.NoError(t, err)

			require.Equal(t, "/admin/realms/realm/organizations/org-1/members", gotPath)
			require.Equal(t, tc.wantQuery, gotQuery)
			require.Len(t, members, 1)
			require.Equal(t, []string{"UPDATE_PASSWORD"}, *members[0].RequiredActions)
		})
	}
}

func TestInviteUserToOrganizationByEmail_FormData(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		params   gocloak.InviteeFormParams
		wantForm url.Values
	}{
		"nil names are omitted": {
			params:   gocloak.InviteeFormParams{Email: gocloak.StringP("jane@example.com")},
			wantForm: url.Values{"email": {"jane@example.com"}},
		},
		"only first name": {
			params: gocloak.InviteeFormParams{
				Email:     gocloak.StringP("jane@example.com"),
				FirstName: gocloak.StringP("Jane"),
			},
			wantForm: url.Values{"email": {"jane@example.com"}, "firstName": {"Jane"}},
		},
		"both names": {
			params: gocloak.InviteeFormParams{
				Email:     gocloak.StringP("jane@example.com"),
				FirstName: gocloak.StringP("Jane"),
				LastName:  gocloak.StringP("Doe"),
			},
			wantForm: url.Values{"email": {"jane@example.com"}, "firstName": {"Jane"}, "lastName": {"Doe"}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var gotMethod, gotPath, gotContentType string
			var gotForm url.Values
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				gotContentType = r.Header.Get("Content-Type")
				if err := r.ParseForm(); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				gotForm = r.PostForm
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := gocloak.NewClient(server.URL)
			err := client.InviteUserToOrganizationByEmail(context.Background(), "token", "realm", "org-1", tc.params)
			require.NoError(t, err)

			require.Equal(t, http.MethodPost, gotMethod)
			require.Equal(t, "/admin/realms/realm/organizations/org-1/members/invite-user", gotPath)
			require.Equal(t, "application/x-www-form-urlencoded", gotContentType)
			require.Equal(t, tc.wantForm, gotForm)
		})
	}
}

func TestInviteUserToOrganizationByEmail_RequiresEmail(t *testing.T) {
	t.Parallel()

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := gocloak.NewClient(server.URL)
	err := client.InviteUserToOrganizationByEmail(context.Background(), "token", "realm", "org-1", gocloak.InviteeFormParams{})
	require.Error(t, err)
	require.False(t, called)
}
