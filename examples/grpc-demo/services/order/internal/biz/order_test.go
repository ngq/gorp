package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type orderRepoStub struct {
	created *Order
	orders  []*Order
}

func (s *orderRepoStub) Create(ctx context.Context, order *Order) error {
	s.created = order
	order.ID = 1
	return nil
}

func (s *orderRepoStub) GetByID(ctx context.Context, id uint) (*Order, error) {
	return &Order{ID: id, UserID: 2, ProductID: 3, ProductName: "book", Quantity: 1, TotalPrice: 19.9, Status: "pending"}, nil
}

func (s *orderRepoStub) List(ctx context.Context, page, size int) ([]*Order, int64, error) {
	return s.orders, int64(len(s.orders)), nil
}

func (s *orderRepoStub) ListByUserID(ctx context.Context, userID uint, page, size int) ([]*Order, int64, error) {
	return s.orders, int64(len(s.orders)), nil
}

func (s *orderRepoStub) Delete(ctx context.Context, id uint) error {
	return nil
}

func TestOrderUseCaseCreate(t *testing.T) {
	repo := &orderRepoStub{}
	uc := NewOrderUseCase(repo)

	order, err := uc.Create(context.Background(), 2, 3, "book", 1, 19.9)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, uint(1), order.ID)
	require.Equal(t, "pending", repo.created.Status)
}

func TestOrderUseCaseList(t *testing.T) {
	repo := &orderRepoStub{orders: []*Order{
		{ID: 1, ProductName: "book", Status: "pending"},
	}}
	uc := NewOrderUseCase(repo)

	items, total, err := uc.List(context.Background(), 1, 10)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), total)
	require.Equal(t, "book", items[0].ProductName)
}
