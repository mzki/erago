package state

import (
	"context"
	"encoding/json"
	"fmt"
)

type StubRepository struct {
	infoBuf   []byte
	systemBuf []byte
	shareBuf  []byte
}

func (r *StubRepository) Exist(ctx context.Context, id int) bool {
	return true
}

func (r *StubRepository) Clear() {
	for _, buf := range []*[]byte{
		&r.infoBuf,
		&r.systemBuf,
		&r.shareBuf,
	} {
		if buf == nil {
			continue
		}
		*buf = (*buf)[:0]
	}
}

func marshalAndStore(dst *[]byte, src interface{}) error {
	buf, err := json.Marshal(src)
	if err != nil {
		return err
	}
	*dst = buf
	return nil
}

func unmarshalAndStore(src []byte, dst interface{}) error {
	if src == nil {
		return fmt.Errorf("nil source buffer for unmarshal")
	}
	if len(src) == 0 {
		return fmt.Errorf("empty source buffer for unmarshal")
	}
	return json.Unmarshal(src, dst)
}

func (r *StubRepository) SaveSystemData(ctx context.Context, id int, state *SystemData, info *SaveInfo) error {
	if err := marshalAndStore(&r.infoBuf, info); err != nil {
		return err
	}
	return marshalAndStore(&r.systemBuf, state)
}

func (r *StubRepository) LoadSystemData(ctx context.Context, id int, state *SystemData, info *SaveInfo) error {
	if err := unmarshalAndStore(r.infoBuf, info); err != nil {
		return err
	}
	return unmarshalAndStore(r.systemBuf, state)
}

func (r *StubRepository) SaveShareData(ctx context.Context, state *UserVariables) error {
	return marshalAndStore(&r.shareBuf, state)
}

func (r *StubRepository) LoadShareData(ctx context.Context, state *UserVariables) error {
	return unmarshalAndStore(r.shareBuf, state)
}

func (r StubRepository) LoadMetaList(ctx context.Context, ids ...int) ([]*MetaData, error) {
	return nil, nil
}
