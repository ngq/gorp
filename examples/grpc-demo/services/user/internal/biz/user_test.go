package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type userRepoStub struct {
	created *User
	users   []*User
}

func (s *userRepoStub) Create(ctx context.Context, user *User) error {
	s.created = user
	user.ID = 1
	return nil
}

func (s *userRepoStub) GetByID(ctx context.Context, id uint) (*User, error) {
	return &User{ID: id, Username: "alice", Email: "alice@example.com"}, nil
}

func (s *userRepoStub) GetByUsername(ctx context.Context, username string) (*User, error) {
	return &User{ID: 1, Username: username, Email: username + "@example.com"}, nil
}

func (s *userRepoStub) List(ctx context.Context, page, size int) ([]*User, int64, error) {
	return s.users, int64(len(s.users)), nil
}

func (s *userRepoStub) Delete(ctx context.Context, id uint) error {
	return nil
}

func TestUserUseCaseCreate(t *testing.T) {
	repo := &userRepoStub{}
	uc := NewUserUseCase(repo)

	user, err := uc.Create(context.Background(), "alice", "alice@example.com")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, uint(1), user.ID)
	require.Equal(t, "alice", repo.created.Username)
	require.Equal(t, "alice@example.com", repo.created.Email)
}

func TestUserUseCaseList(t *testing.T) {
	repo := &userRepoStub{users: []*User{
		{ID: 1, Username: "alice", Email: "alice@example.com"},
	}}
	uc := NewUserUseCase(repo)

	items, total, err := uc.List(context.Background(), 1, 10)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), total)
	require.Equal(t, "alice", items[0].Username)
}
