package service

import (
	"encoding/json"
	"testing"

	"gpt-image-playground/backend/database"
)

func TestNormalizeTaskN(t *testing.T) {
	cases := []struct {
		name string
		in   int
		want int
	}{
		{name: "negative", in: -5, want: 1},
		{name: "zero", in: 0, want: 1},
		{name: "one", in: 1, want: 1},
		{name: "max", in: MaxTaskN, want: MaxTaskN},
		{name: "above max", in: MaxTaskN + 1, want: MaxTaskN},
		{name: "huge", in: 999, want: MaxTaskN},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeTaskN(tc.in); got != tc.want {
				t.Fatalf("NormalizeTaskN(%d) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestAuthUserJSON_IncludesUsername(t *testing.T) {
	au := AuthUser{
		ID:       "user-1",
		Label:    "testlabel",
		Username: "cooluser",
		Role:     "user",
		Quota:    10,
	}
	data, err := json.Marshal(au)
	if err != nil {
		t.Fatalf("marshal AuthUser: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal AuthUser: %v", err)
	}
	if v, ok := parsed["username"]; !ok || v != "cooluser" {
		t.Errorf("username field: ok=%v val=%v, want 'cooluser'", ok, v)
	}
}

func TestAuthUserJSON_OmitsUsernameWhenEmpty(t *testing.T) {
	au := AuthUser{
		ID:    "user-1",
		Label: "testlabel",
		Role:  "user",
	}
	data, err := json.Marshal(au)
	if err != nil {
		t.Fatalf("marshal AuthUser: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal AuthUser: %v", err)
	}
	if _, ok := parsed["username"]; ok {
		t.Error("username should be omitted when empty string")
	}
}

func TestAuthUserJSON_IncludesNeedsMigrationWhenTrue(t *testing.T) {
	au := AuthUser{
		ID:             "user-1",
		Label:          "migrate-me",
		Role:           "user",
		NeedsMigration: true,
	}
	data, err := json.Marshal(au)
	if err != nil {
		t.Fatalf("marshal AuthUser: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal AuthUser: %v", err)
	}
	if v, ok := parsed["needsMigration"]; !ok || v != true {
		t.Errorf("needsMigration field: ok=%v val=%v, want true", ok, v)
	}
}

func TestAuthUserJSON_OmitsNeedsMigrationWhenFalse(t *testing.T) {
	au := AuthUser{
		ID:    "user-1",
		Label: "no-migrate",
		Role:  "user",
	}
	data, err := json.Marshal(au)
	if err != nil {
		t.Fatalf("marshal AuthUser: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal AuthUser: %v", err)
	}
	if _, ok := parsed["needsMigration"]; ok {
		t.Error("needsMigration should be omitted when false")
	}
}

func TestAdminUserJSON_IncludesUsername(t *testing.T) {
	au := AdminUser{
		ID:       "admin-1",
		Label:    "admin",
		Username: "theadmin",
		Role:     "admin",
		Status:   "active",
	}
	data, err := json.Marshal(au)
	if err != nil {
		t.Fatalf("marshal AdminUser: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal AdminUser: %v", err)
	}
	if v, ok := parsed["username"]; !ok || v != "theadmin" {
		t.Errorf("username field: ok=%v val=%v, want 'theadmin'", ok, v)
	}
}

func TestUserDTO_PasswordHashNeverSerialized(t *testing.T) {
	ph := "secret123"
	u := User{
		ID:           "u1",
		Label:        "test",
		PasswordHash: &ph,
	}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal User: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal User: %v", err)
	}
	if _, ok := parsed["passwordHash"]; ok {
		t.Error("passwordHash must never appear in JSON output")
	}
}

func TestUserDTO_InviteCodeSerialized(t *testing.T) {
	ic := "MYCODE"
	u := User{
		ID:         "u1",
		Label:      "test",
		InviteCode: &ic,
	}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal User: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal User: %v", err)
	}
	if v, ok := parsed["inviteCode"]; !ok || v != "MYCODE" {
		t.Errorf("inviteCode field: ok=%v val=%v, want 'MYCODE'", ok, v)
	}
}

func TestUserDTO_InviteCodeSetAtNeverSerialized(t *testing.T) {
	ts := int64(1716451200000)
	u := User{
		ID:              "u1",
		Label:           "test",
		InviteCodeSetAt: &ts,
	}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal User: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal User: %v", err)
	}
	if _, ok := parsed["inviteCodeSetAt"]; ok {
		t.Error("inviteCodeSetAt must never appear in JSON output")
	}
}

// TestDbUserToAuthUser tests the helper that converts a database.User to a
// service.AuthUser with the correct NeedsMigration and Username logic.
func TestDbUserToAuthUser_NeedsMigrationTrueWhenNoPasswordHash(t *testing.T) {
	username := "nooby"
	u := &database.User{
		ID:       "u-needs-mig",
		Label:    "nooby-label",
		Role:     "user",
		Quota:    10,
		UsedCount: 2,
		Username:  &username,
		// PasswordHash is nil — old user.
	}
	au := dbUserToAuthUser(u)
	if !au.NeedsMigration {
		t.Error("NeedsMigration should be true when PasswordHash is nil")
	}
	if au.Username != "nooby" {
		t.Errorf("Username = %q, want 'nooby'", au.Username)
	}
	if au.ID != "u-needs-mig" {
		t.Errorf("ID = %q, want 'u-needs-mig'", au.ID)
	}
	if au.Label != "nooby-label" {
		t.Errorf("Label = %q, want 'nooby-label'", au.Label)
	}
}

func TestDbUserToAuthUser_NeedsMigrationFalseWhenPasswordHashSet(t *testing.T) {
	username := "pro"
	ph := "$2a$10$something"
	u := &database.User{
		ID:           "u-migrated",
		Label:        "pro-label",
		Role:         "user",
		Quota:        20,
		UsedCount:    5,
		Username:     &username,
		PasswordHash: &ph,
	}
	au := dbUserToAuthUser(u)
	if au.NeedsMigration {
		t.Error("NeedsMigration should be false when PasswordHash is set")
	}
	if au.Username != "pro" {
		t.Errorf("Username = %q, want 'pro'", au.Username)
	}
}

func TestDbUserToAuthUser_UsernameEmptyWhenNil(t *testing.T) {
	u := &database.User{
		ID:    "u-no-uname",
		Label: "nolabel",
		Role:  "user",
		// Username is nil.
	}
	au := dbUserToAuthUser(u)
	if au.Username != "" {
		t.Errorf("Username = %q, want empty string for nil Username", au.Username)
	}
}
