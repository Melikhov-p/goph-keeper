package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Melikhov-p/goph-keeper/internal/domain/user"
)

// mockUserRepo реализует Repository для тестирования
type mockUserRepo struct {
	createFunc     func(ctx context.Context, user *user.User) error
	getByIDFunc    func(ctx context.Context, id int) (*user.User, error)
	getByLoginFunc func(ctx context.Context, login string) (*user.User, error)
	updateFunc     func(ctx context.Context, user *user.User) error
	deleteFunc     func(ctx context.Context, id int) error
}

func (m *mockUserRepo) Create(ctx context.Context, user *user.User) error {
	return m.createFunc(ctx, user)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int) (*user.User, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockUserRepo) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	return m.getByLoginFunc(ctx, login)
}

func (m *mockUserRepo) Update(ctx context.Context, user *user.User) error {
	return m.updateFunc(ctx, user)
}

func (m *mockUserRepo) Delete(ctx context.Context, id int) error {
	return m.deleteFunc(ctx, id)
}

func TestService_Register(t *testing.T) {
	tests := []struct {
		name      string
		repoSetup func() *mockUserRepo
		login     string
		password  string
		pepper    string
		wantUser  bool
		wantErr   error
	}{
		{
			name: "successful registration",
			repoSetup: func() *mockUserRepo {
				return &mockUserRepo{
					getByLoginFunc: func(ctx context.Context, login string) (*user.User, error) {
						return nil, user.ErrNotFound
					},
					createFunc: func(ctx context.Context, user *user.User) error {
						return nil
					},
				}
			},
			login:    "newuser",
			password: "secure123",
			pepper:   "pepper",
			wantUser: true,
		},
		{
			name: "user already exists",
			repoSetup: func() *mockUserRepo {
				return &mockUserRepo{
					getByLoginFunc: func(ctx context.Context, login string) (*user.User, error) {
						return &user.User{ID: 1, Login: "existing"}, nil
					},
				}
			},
			login:    "existing",
			password: "password",
			pepper:   "pepper",
			wantErr:  user.ErrAlreadyExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := user.NewService(tt.repoSetup())
			user, err := s.Register(context.Background(), tt.login, tt.password, tt.pepper)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) && !contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("unexpected error: got %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.wantUser && user != nil {
				t.Error("expected nil user, got user")
			}

			if tt.wantUser && user == nil {
				t.Error("expected user, got nil")
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	validUser, _ := user.NewUser("valid", "password", "pepper")
	validUser.ID = 1

	tests := []struct {
		name      string
		repoSetup func() *mockUserRepo
		login     string
		password  string
		pepper    string
		wantUser  bool
		wantErr   error
	}{
		{
			name: "successful login",
			repoSetup: func() *mockUserRepo {
				return &mockUserRepo{
					getByLoginFunc: func(ctx context.Context, login string) (*user.User, error) {
						return validUser, nil
					},
				}
			},
			login:    "valid",
			password: "password",
			pepper:   "pepper",
			wantUser: true,
		},
		{
			name: "user not found",
			repoSetup: func() *mockUserRepo {
				return &mockUserRepo{
					getByLoginFunc: func(ctx context.Context, login string) (*user.User, error) {
						return nil, user.ErrNotFound
					},
				}
			},
			login:    "nonexistent",
			password: "password",
			pepper:   "pepper",
			wantErr:  user.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := user.NewService(tt.repoSetup())
			user, err := s.Login(context.Background(), tt.login, tt.password, tt.pepper)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) && !contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("unexpected error: got %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.wantUser && user != nil {
				t.Error("expected nil user, got user")
			}

			if tt.wantUser && user == nil {
				t.Error("expected user, got nil")
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	existingUser := &user.User{ID: 1, Login: "existing"}

	tests := []struct {
		name      string
		repoSetup func() *mockUserRepo
		user      *user.User
		wantErr   error
	}{
		{
			name: "successful update",
			repoSetup: func() *mockUserRepo {
				return &mockUserRepo{
					getByIDFunc: func(ctx context.Context, id int) (*user.User, error) {
						return existingUser, nil
					},
					updateFunc: func(ctx context.Context, user *user.User) error {
						return nil
					},
				}
			},
			user: &user.User{ID: 1, Login: "updated"},
		},
		{
			name: "user not found",
			repoSetup: func() *mockUserRepo {
				return &mockUserRepo{
					getByIDFunc: func(ctx context.Context, id int) (*user.User, error) {
						return nil, user.ErrNotFound
					},
				}
			},
			user:    &user.User{ID: 999, Login: "nonexistent"},
			wantErr: user.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := user.NewService(tt.repoSetup())
			err := s.Update(context.Background(), tt.user)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("unexpected error: got %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestService_GetUserByID(t *testing.T) {
	existingUser := &user.User{ID: 1, Login: "existing"}

	tests := []struct {
		name      string
		repoSetup func() *mockUserRepo
		userID    int
		wantUser  bool
		wantErr   error
	}{
		{
			name: "successful get",
			repoSetup: func() *mockUserRepo {
				return &mockUserRepo{
					getByIDFunc: func(ctx context.Context, id int) (*user.User, error) {
						return existingUser, nil
					},
				}
			},
			userID:   1,
			wantUser: true,
		},
		{
			name: "user not found",
			repoSetup: func() *mockUserRepo {
				return &mockUserRepo{
					getByIDFunc: func(ctx context.Context, id int) (*user.User, error) {
						return nil, user.ErrNotFound
					},
				}
			},
			userID:  999,
			wantErr: user.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := user.NewService(tt.repoSetup())
			user, err := s.GetUserByID(context.Background(), tt.userID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) && !contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("unexpected error: got %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.wantUser && user != nil {
				t.Error("expected nil user, got user")
			}

			if tt.wantUser && user == nil {
				t.Error("expected user, got nil")
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
