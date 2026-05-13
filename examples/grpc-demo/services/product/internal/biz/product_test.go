package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type productRepoStub struct {
	created  *Product
	products []*Product
}

func (s *productRepoStub) Create(ctx context.Context, product *Product) error {
	s.created = product
	product.ID = 1
	return nil
}

func (s *productRepoStub) GetByID(ctx context.Context, id uint) (*Product, error) {
	return &Product{ID: id, Name: "book", Description: "demo", Price: 19.9, Stock: 10}, nil
}

func (s *productRepoStub) List(ctx context.Context, page, size int) ([]*Product, int64, error) {
	return s.products, int64(len(s.products)), nil
}

func (s *productRepoStub) Delete(ctx context.Context, id uint) error {
	return nil
}

func TestProductUseCaseCreate(t *testing.T) {
	repo := &productRepoStub{}
	uc := NewProductUseCase(repo)

	product, err := uc.Create(context.Background(), "book", "demo", 19.9, 10)
	require.NoError(t, err)
	require.NotNil(t, product)
	require.Equal(t, uint(1), product.ID)
	require.Equal(t, "book", repo.created.Name)
}

func TestProductUseCaseList(t *testing.T) {
	repo := &productRepoStub{products: []*Product{
		{ID: 1, Name: "book", Price: 19.9},
	}}
	uc := NewProductUseCase(repo)

	items, total, err := uc.List(context.Background(), 1, 10)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), total)
	require.Equal(t, "book", items[0].Name)
}
