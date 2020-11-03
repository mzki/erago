package state

import (
	"context"
)

type StubRepository struct{}

func (r StubRepository) Exist(ctx context.Context, id int) bool {
	return true
}

func (r StubRepository) SaveSystemData(ctx context.Context, id int, state *SystemData, info *SaveInfo) error {
	return nil
}

func (r StubRepository) LoadSystemData(ctx context.Context, id int, state *SystemData, info *SaveInfo) error {
	return nil
}

func (r StubRepository) SaveShareData(ctx context.Context, state *UserVariables) error {
	return nil
}

func (r StubRepository) LoadShareData(ctx context.Context, state *UserVariables) error {
	return nil
}

func (r StubRepository) LoadMetaList(ctx context.Context, ids ...int) ([]*MetaData, error) {
	return nil, nil
}
