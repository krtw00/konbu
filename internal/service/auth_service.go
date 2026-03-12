package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/apperror"
	"github.com/krtw00/konbu/internal/config"
	"github.com/krtw00/konbu/internal/model"
	"github.com/krtw00/konbu/internal/repository"
)

type AuthService struct {
	queries *repository.Queries
	db      *sql.DB
	cfg     *config.Config
}

func NewAuthService(db *sql.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		queries: repository.New(db),
		db:      db,
		cfg:     cfg,
	}
}

func (s *AuthService) GetOrCreateUser(ctx context.Context, email string) (*model.User, error) {
	if !s.cfg.IsEmailAllowed(email) {
		return nil, apperror.Forbidden("email not allowed")
	}

	u, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil {
		return toModelUser(u), nil
	}
	if !errors.Is(err, repository.ErrNoRows) {
		return nil, apperror.Internal(err)
	}

	count, err := s.queries.CountUsers(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	isAdmin := count == 0
	if s.cfg.AdminEmail != "" && strings.EqualFold(email, s.cfg.AdminEmail) {
		isAdmin = true
	}

	u, err = s.queries.CreateUser(ctx, email, "", isAdmin)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return toModelUser(u), nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	u, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("user")
		}
		return nil, apperror.Internal(err)
	}
	return toModelUser(u), nil
}

func (s *AuthService) UpdateUser(ctx context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.User, error) {
	u, err := s.queries.UpdateUser(ctx, id, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.NotFound("user")
		}
		return nil, apperror.Internal(err)
	}
	return toModelUser(u), nil
}

func (s *AuthService) AuthenticateByAPIKey(ctx context.Context, rawKey string) (*model.User, error) {
	hash := hashAPIKey(rawKey)
	ak, err := s.queries.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.Unauthorized("invalid api key")
		}
		return nil, apperror.Internal(err)
	}

	_ = s.queries.UpdateAPIKeyLastUsed(ctx, ak.ID)

	return &model.User{
		ID:      ak.UserID,
		Email:   ak.Email,
		Name:    ak.UserName,
		IsAdmin: ak.IsAdmin,
	}, nil
}

func (s *AuthService) CreateAPIKey(ctx context.Context, userID uuid.UUID, req model.CreateAPIKeyRequest) (*model.APIKey, error) {
	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, apperror.Internal(err)
	}

	hash := hashAPIKey(rawKey)
	ak, err := s.queries.CreateAPIKey(ctx, userID, req.Name, hash)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	return &model.APIKey{
		ID:        ak.ID,
		Name:      ak.Name,
		Key:       rawKey,
		CreatedAt: ak.CreatedAt,
	}, nil
}

func (s *AuthService) ListAPIKeys(ctx context.Context, userID uuid.UUID) ([]model.APIKey, error) {
	keys, err := s.queries.ListAPIKeysByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	result := make([]model.APIKey, len(keys))
	for i, k := range keys {
		result[i] = model.APIKey{
			ID:         k.ID,
			Name:       k.Name,
			LastUsedAt: k.LastUsedAt,
			CreatedAt:  k.CreatedAt,
		}
	}
	return result, nil
}

func (s *AuthService) DeleteAPIKey(ctx context.Context, id, userID uuid.UUID) error {
	return s.queries.DeleteAPIKey(ctx, id, userID)
}

func toModelUser(u repository.User) *model.User {
	return &model.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func generateAPIKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("konbu_%s", hex.EncodeToString(b)), nil
}

func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
