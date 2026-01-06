package session

import (
	"context"

	"github.com/goodieshq/onus/internal/server/core"
	"github.com/google/uuid"
	"github.com/oklog/ulid"
)

type ctxKey string

const ctxKeyUID ctxKey = "user_id"

// CtxSetUID sets the session owner's user ID in the context
func CtxSetUID(ctx context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKeyUID, uid)
}

// CtxGetUID gets the session owner's user ID from the context
func CtxGetUID(ctx context.Context) uuid.UUID {
	uid, ok := ctx.Value(ctxKeyUID).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return uid
}

const ctxKeyOID ctxKey = "organization_id"

// CtxSetOID sets the session owner's organization ID in the context
func CtxSetOID(ctx context.Context, oid uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKeyOID, oid)
}

// CtxGetOID gets the session owner's organization ID from the context
func CtxGetOID(ctx context.Context) uuid.UUID {
	oid, ok := ctx.Value(ctxKeyOID).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return oid
}

const ctxKeyRole ctxKey = "role"

// CtxSetRole sets the session owner's role in the context
func CtxSetRole(ctx context.Context, role core.UserOrgRole) context.Context {
	return context.WithValue(ctx, ctxKeyRole, role)
}

// CtxGetRole gets the session owner's role from the context
func CtxGetRole(ctx context.Context) core.UserOrgRole {
	role, ok := ctx.Value(ctxKeyRole).(core.UserOrgRole)
	if !ok || !role.IsValid() {
		return core.RoleNone
	}
	return role
}

const ctxKeyPathTaskID ctxKey = "param_task_id"

// CtxSetPathTaskID sets the task ID from the URL path in the context .../{task_id}/...
func CtxSetPathTaskID(ctx context.Context, taskID ulid.ULID) context.Context {
	return context.WithValue(ctx, ctxKeyPathTaskID, taskID)
}

// CtxGetPathTaskID gets the task ID from the URL path in the context .../{task_id}/...
func CtxGetPathTaskID(ctx context.Context) ulid.ULID {
	taskID, ok := ctx.Value(ctxKeyPathTaskID).(ulid.ULID)
	if !ok {
		return ulid.ULID{}
	}
	return taskID
}

const ctxKeyPathUserID ctxKey = "param_user_id"

// CtxSetPathUserID sets the user ID from the URL path in the context .../{user_id}/...
func CtxSetPathUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKeyPathUserID, userID)
}

// CtxGetPathUserID gets the user ID from the URL path in the context .../{user_id}/...
func CtxGetPathUserID(ctx context.Context) uuid.UUID {
	userID, ok := ctx.Value(ctxKeyPathUserID).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return userID
}

const ctxKeyParamPretty ctxKey = "param_pretty"

// CtxSetParamPretty sets the "pretty" query parameter in the context ?pretty=...
func CtxSetParamPretty(ctx context.Context, pretty bool) context.Context {
	return context.WithValue(ctx, ctxKeyParamPretty, pretty)
}

// CtxGetParamPretty gets the "pretty" query parameter from the context ?pretty=...
func CtxGetParamPretty(ctx context.Context) bool {
	pretty, ok := ctx.Value(ctxKeyParamPretty).(bool)
	if !ok {
		return false
	}
	return pretty
}
