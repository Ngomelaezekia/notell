package main

import "testing"

func TestC4Roles(t *testing.T) {
	for _, role := range []Role{RoleAdmin, RoleEditor, RoleBroadcaster, RoleModerator, RoleAnalyst} {
		if !validRole(role) { t.Fatalf("expected %s to be a valid team role", role) }
	}
	if validRole(RoleOwner) { t.Fatal("owner should not be assignable through team invitations") }
	if !rolePermissions[RoleAdmin]["team"] { t.Fatal("admin must manage team") }
	if rolePermissions[RoleAnalyst]["content"] { t.Fatal("analyst must not manage content") }
}

func TestC5StatusValues(t *testing.T) {
	p := ChannelProgram{}
	s := ChannelSchedule{}
	if p.Status != "" { t.Fatal("zero program status should be empty before handler defaulting") }
	if s.Status != "" { t.Fatal("zero schedule status should be empty before handler defaulting") }
}
