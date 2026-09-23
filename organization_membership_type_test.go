package gocloak

import (
	"encoding/json"
	"testing"
)

// Regression test for the previously broken MembershipType (empty struct + a
// misspelled `membershipetype` json tag) which silently dropped the value
// Keycloak returns under `membershipType`.
func TestMemberRepresentationMembershipTypeUnmarshal(t *testing.T) {
	cases := map[string]MembershipType{
		`{"id":"u1","membershipType":"MANAGED"}`:   MembershipTypeManaged,
		`{"id":"u2","membershipType":"UNMANAGED"}`: MembershipTypeUnmanaged,
	}
	for body, want := range cases {
		var m MemberRepresentation
		err := json.Unmarshal([]byte(body), &m)
		if err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		if m.MembershipType == nil {
			t.Fatalf("membershipType was dropped for %s", body)
		}
		if *m.MembershipType != want {
			t.Fatalf("got %q, want %q for %s", *m.MembershipType, want, body)
		}
	}
}

func TestGetMembersParamsMembershipTypeQuery(t *testing.T) {
	mt := MembershipTypeManaged
	params, err := GetQueryParams(GetMembersParams{MembershipType: &mt})
	if err != nil {
		t.Fatalf("GetQueryParams: %v", err)
	}
	if params["membershipType"] != "MANAGED" {
		t.Fatalf("query param membershipType=%q, want MANAGED (got params %v)", params["membershipType"], params)
	}
}
