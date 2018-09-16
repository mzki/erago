package state

import "context"

// Repository is a abstract data-store which
// persists GameState to arbitrary storage.
type Repository interface {

	// Exist returns whether a persisted GameState with id exists?
	// return true if exists. return false if not exists or
	// context is canceled.
	Exist(ctx context.Context, id int) bool

	// SaveSystemData persists SystemData, the game specific data,
	// in the given GameState.
	// ID is used to identify the persisted data.
	// return nil on success. return some error on failure or
	// context is canceled.
	SaveSystemData(ctx context.Context, id int, state *GameState) error

	// LoadSystemData restores SystemData, the game specific data,
	// into the given GameState.
	// ID is used to identify the persisted data.
	// return nil on success. return some error on failure or
	// context is canceled.
	LoadSystemData(ctx context.Context, id int, state *GameState) error

	// SaveSharedData persists ShareData, sharing over different GameState,
	// in the given GameState. return nil on success. return some error
	// on failure or context is canceled.
	SaveShareData(ctx context.Context, state *GameState) error

	// LoadShareData restores ShareData, sharing over different GameState,
	// into the given GameState.
	// return nil on success. return some error on failure or
	// context is canceled.
	LoadShareData(ctx context.Context, state *GameState) error

	// LoadMetaList returns list of meta data associated to id list.
	// If id list is empty or nil, return empty list of meta data for persisted
	// GameState.
	// It also returns error if any id is not found or context is canceled.
	LoadMetaList(ctx context.Context, ids ...int) ([]*FileHeader, error)
}
