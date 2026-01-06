package session

import (
	"context"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/goodieshq/onus/internal/server/core"
)

const AuthSnapshotTTL = 30 * time.Second

// InvalidateSnapshot marks the auth snapshot as invalid by resetting its last update time
func InvalidateSnapshot(ctx context.Context, sm *scs.SessionManager) {
	sm.Put(ctx, "auth_snapshot_last_update", time.Time{})
}

// Outdated checks if the auth snapshot is outdated based on the TTL
func Outdated(ctx context.Context, sm *scs.SessionManager) bool {
	lastUpdate := sm.GetTime(ctx, "auth_snapshot_last_update")
	if lastUpdate.IsZero() {
		return true
	}
	return time.Since(lastUpdate) > AuthSnapshotTTL
}

// SetAuthSnapshotIfDifferent updates the auth snapshot in the session if it has changed
func SetAuthSnapshotIfDifferent(ctx context.Context, sm *scs.SessionManager, snapshotOld *core.AuthSnapshot, snapshotNew *core.AuthSnapshot) {
	if snapshotNew == nil {
		return
	}

	if snapshotOld == nil {
		SetAuthSnapshot(ctx, sm, snapshotNew)
	}

	if snapshotOld.Role != snapshotNew.Role ||
		snapshotOld.OrganizationName != snapshotNew.OrganizationName ||
		snapshotOld.Disabled != snapshotNew.Disabled ||
		snapshotOld.Name != snapshotNew.Name {
		SetAuthSnapshot(ctx, sm, snapshotNew)
	}
}

// SetAuthSnapshot stores the auth snapshot details in the session
func SetAuthSnapshot(ctx context.Context, sm *scs.SessionManager, snapshot *core.AuthSnapshot) {
	sm.Put(ctx, "auth_snapshot_last_update", time.Now())
	sm.Put(ctx, "user_id", snapshot.UserID)
	sm.Put(ctx, "email", snapshot.Email)
	sm.Put(ctx, "name", snapshot.Name)
	sm.Put(ctx, "organization_id", snapshot.OrganizationID)
	sm.Put(ctx, "organization_name", snapshot.OrganizationName)
	sm.Put(ctx, "organization_domain", snapshot.OrganizationDomain)
	sm.Put(ctx, "role", int(snapshot.Role))
	sm.Put(ctx, "disabled", snapshot.Disabled)
}

// GetAuthSnapshot retrieves the auth snapshot details from the session
func GetAuthSnapshot(ctx context.Context, sm *scs.SessionManager) (*core.AuthSnapshot, error) {
	return &core.AuthSnapshot{
		UserID:             sm.GetString(ctx, "user_id"),
		Email:              sm.GetString(ctx, "email"),
		Name:               sm.GetString(ctx, "name"),
		OrganizationID:     sm.GetString(ctx, "organization_id"),
		OrganizationName:   sm.GetString(ctx, "organization_name"),
		OrganizationDomain: sm.GetString(ctx, "organization_domain"),
		Role:               core.UserOrgRole(sm.GetInt(ctx, "role")),
		Disabled:           sm.GetBool(ctx, "disabled"),
	}, nil
}

// ClearSession destroys the current session and loges the user out
func ClearSession(ctx context.Context, sm *scs.SessionManager) error {
	return sm.Destroy(ctx)
}
