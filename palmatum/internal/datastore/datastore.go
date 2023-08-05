package datastore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var ErrNotFound = errors.New("datastore: key not found")

type Datastore struct {
	lock     sync.Mutex
	filename string
	cached   map[string]json.RawMessage
}

func New(filename string) *Datastore {
	ds := &Datastore{
		filename: filename,
	}

	return ds
}

func (ds *Datastore) load() error {
	dat, err := os.ReadFile(ds.filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ds.cached = make(map[string]json.RawMessage)
			return nil
		}
		return fmt.Errorf("loading datastore %#v from disk: %w", ds.filename, err)
	}
	if err := json.Unmarshal(dat, &ds.cached); err != nil {
		return fmt.Errorf("unmarshaling datastore JSON from %#v: %w", ds.filename, err)
	}
	return nil
}

func (ds *Datastore) dump() error {
	dat, err := json.Marshal(ds.cached)
	if err != nil {
		return fmt.Errorf("marshaling datastore to JSON: %w", err)
	}
	if err := os.WriteFile(ds.filename, dat, 0644); err != nil {
		return fmt.Errorf("writing datastore JSON to %#v: %w", ds.filename, err)
	}
	return nil
}

func (ds *Datastore) Get(key string, output any) error {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	if ds.cached == nil {
		if err := ds.load(); err != nil {
			return err
		}
	}

	v, found := ds.cached[key]
	if !found {
		return ErrNotFound
	}

	return json.Unmarshal(v, output)
}

func (ds *Datastore) Put(key string, data any) error {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	if ds.cached == nil {
		if err := ds.load(); err != nil {
			return err
		}
	}

	dat, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling object to JSON: %w", err)
	}

	ds.cached[key] = json.RawMessage(dat)

	return ds.dump()
}
