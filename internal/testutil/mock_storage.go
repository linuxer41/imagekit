package testutil

import (
	"context"
	"fmt"
)

type MockStorage struct {
	Data map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{Data: make(map[string][]byte)}
}

func (m *MockStorage) Get(_ context.Context, path string) ([]byte, error) {
	d, ok := m.Data[path]
	if !ok {
		return nil, fmt.Errorf("not found: %s", path)
	}
	return d, nil
}

func (m *MockStorage) Delete(_ context.Context, path string) error {
	if _, ok := m.Data[path]; !ok {
		return fmt.Errorf("not found: %s", path)
	}
	delete(m.Data, path)
	return nil
}

func (m *MockStorage) TestConnection(_ context.Context) error {
	return nil
}
