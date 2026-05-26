package actions

import (
	"reflect"
	"testing"
)

// fakeAction lets us drive MethodFromAlias / PrimaryAliases without
// touching real action dispatch logic.
type fakeAction struct{}

func (fakeAction) Default() string { return "list" }
func (fakeAction) Commands() map[string]string {
	return map[string]string{
		"list":   "list, l",
		"create": "create",
		"delete": "delete, del, d",
	}
}
func (fakeAction) RequireGit() bool                          { return false }
func (fakeAction) Dispatch(method string, args []string) error { return nil }

func TestMethodFromAlias(t *testing.T) {
	a := fakeAction{}
	tests := []struct {
		alias string
		want  string
	}{
		{"list", "list"},
		{"l", "list"},
		{"create", "create"},
		{"delete", "delete"},
		{"del", "delete"},
		{"d", "delete"},
		{"missing", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := MethodFromAlias(a, tt.alias); got != tt.want {
			t.Errorf("MethodFromAlias(%q) = %q, want %q", tt.alias, got, tt.want)
		}
	}
}

func TestPrimaryAliases(t *testing.T) {
	got := PrimaryAliases(fakeAction{})
	want := []string{"create", "delete", "list"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PrimaryAliases() = %v, want %v", got, want)
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"yes", false}, // strconv.ParseBool only accepts canonical forms
		{"", false},
	}
	for _, tt := range tests {
		if got := parseBool(tt.in); got != tt.want {
			t.Errorf("parseBool(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestYesNo(t *testing.T) {
	if yesNo(true) != "Yes" {
		t.Error("yesNo(true) should return Yes")
	}
	if yesNo(false) != "No" {
		t.Error("yesNo(false) should return No")
	}
}

func TestRegisteredActionsHaveSaneShape(t *testing.T) {
	// Smoke-test every real action so a registration regression (missing
	// default method, missing alias for it) is caught.
	all := []Action{
		Auth{}, Branch{}, Browse{}, Env{}, Pipeline{}, Pr{}, PrDetails{},
		Upgrade{},
	}
	for _, a := range all {
		if a.Default() == "" {
			t.Errorf("%T: Default() returned empty", a)
		}
		if len(a.Commands()) == 0 {
			t.Errorf("%T: Commands() is empty", a)
		}
		if _, ok := a.Commands()[a.Default()]; !ok {
			t.Errorf("%T: Default() = %q has no entry in Commands()", a, a.Default())
		}
	}
}
