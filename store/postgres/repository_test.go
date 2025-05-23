package postgres

import (
	"errors"
	"github.com/evanebb/regauth/repository"
	"github.com/google/uuid"
	"testing"
	"time"
)

// compareRepositories will check two repository.Repository objects for equality.
// Mostly exists for proper timestamp comparison.
func compareRepositories(r1 repository.Repository, r2 repository.Repository) bool {
	return r1.ID == r2.ID &&
		r1.Namespace == r2.Namespace &&
		r1.Name == r2.Name &&
		r1.Visibility == r2.Visibility &&
		r1.CreatedAt.Equal(r2.CreatedAt)
}

func TestRepositoryStore_GetAllByNamespace(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewRepositoryStore(db)

	repositories, err := s.GetAllByNamespace(t.Context(), "adminuser")
	if err != nil {
		t.Errorf("expected err to be nil, got %q", err)
	}

	if len(repositories) != 2 {
		t.Errorf("expected two repositories, got %d", len(repositories))
	}
}

func TestRepositoryStore_GetByNamespaceAndName(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewRepositoryStore(db)

	t.Run("existing repository", func(t *testing.T) {
		repoID, _ := uuid.Parse("0195cd13-ba14-76fd-b43e-55f190e566bd")
		createdAt, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")

		expectedRepo := repository.Repository{
			ID:         repoID,
			Namespace:  "adminuser",
			Name:       "public-image",
			Visibility: repository.VisibilityPublic,
			CreatedAt:  createdAt,
		}

		repo, err := s.GetByNamespaceAndName(t.Context(), "adminuser", "public-image")
		if err != nil {
			t.Errorf("expected err to be nil, got %q", err)
		}

		if !compareRepositories(expectedRepo, repo) {
			t.Errorf("expected %+v, got %+v", expectedRepo, repo)
		}
	})

	t.Run("repository does not exist", func(t *testing.T) {
		if _, err := s.GetByNamespaceAndName(t.Context(), "foo", "bar"); !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("expected %q, got %q", repository.ErrNotFound, err)
		}
	})
}

func TestRepositoryStore_GetByID(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewRepositoryStore(db)

	t.Run("existing repository", func(t *testing.T) {
		repoID, _ := uuid.Parse("0195cd13-ba14-76fd-b43e-55f190e566bd")
		createdAt, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")

		expectedRepo := repository.Repository{
			ID:         repoID,
			Namespace:  "adminuser",
			Name:       "public-image",
			Visibility: repository.VisibilityPublic,
			CreatedAt:  createdAt,
		}

		repo, err := s.GetByID(t.Context(), repoID)
		if err != nil {
			t.Errorf("expected err to be nil, got %q", err)
		}

		if !compareRepositories(expectedRepo, repo) {
			t.Errorf("expected %+v, got %+v", expectedRepo, repo)
		}
	})

	t.Run("repository does not exist", func(t *testing.T) {
		if _, err := s.GetByID(t.Context(), uuid.Nil); !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("expected %q, got %q", repository.ErrNotFound, err)
		}
	})
}

func TestRepositoryStore_Create(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewRepositoryStore(db)

	repo := repository.Repository{
		ID:         uuid.New(),
		Namespace:  "adminuser",
		Name:       "new-image",
		Visibility: repository.VisibilityPublic,
	}

	t.Run("new repository", func(t *testing.T) {
		if err := s.Create(t.Context(), repo); err != nil {
			t.Errorf("expected nil, got %q", err)
		}
	})

	t.Run("repository already exists", func(t *testing.T) {
		if err := s.Create(t.Context(), repo); !errors.Is(err, repository.ErrAlreadyExists) {
			t.Errorf("expected %q, got %q", repository.ErrAlreadyExists, err)
		}
	})
}

func TestRepositoryStore_DeleteByID(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewRepositoryStore(db)

	repoID, _ := uuid.Parse("0195cd13-ba14-76fd-b43e-55f190e566bd")

	if _, err := s.GetByID(t.Context(), repoID); err != nil {
		t.Errorf("expected repository to exist, got %q", err)
	}

	if err := s.DeleteByID(t.Context(), repoID); err != nil {
		t.Errorf("expected nil, got %q", err)
	}

	if _, err := s.GetByID(t.Context(), repoID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected repository to be deleted, got %q", err)
	}
}
