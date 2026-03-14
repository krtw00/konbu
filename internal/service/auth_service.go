package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

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

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*model.User, error) {
	if email == "" || password == "" {
		return nil, apperror.BadRequest("email and password are required")
	}
	if len(password) < 8 {
		return nil, apperror.BadRequest("password must be at least 8 characters")
	}

	_, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, apperror.BadRequest("email already registered")
	}
	if !errors.Is(err, repository.ErrNoRows) {
		return nil, apperror.Internal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	count, err := s.queries.CountUsers(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	isAdmin := count == 0

	u, err := s.queries.CreateUserWithPassword(ctx, email, name, string(hash), isAdmin)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return toModelUser(u), nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.User, error) {
	if email == "" || password == "" {
		return nil, apperror.Unauthorized("invalid credentials")
	}

	u, err := s.queries.GetUserByEmailWithPassword(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, apperror.Unauthorized("invalid credentials")
		}
		return nil, apperror.Internal(err)
	}

	if u.PasswordHash == nil {
		return nil, apperror.Unauthorized("password not set for this account")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)); err != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}

	return &model.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		IsAdmin:   u.IsAdmin,
		Plan:      u.Plan,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	if newPassword == "" || len(newPassword) < 8 {
		return apperror.BadRequest("new password must be at least 8 characters")
	}

	u, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return apperror.NotFound("user")
		}
		return apperror.Internal(err)
	}

	uwp, err := s.queries.GetUserByEmailWithPassword(ctx, u.Email)
	if err != nil {
		return apperror.Internal(err)
	}

	if uwp.PasswordHash != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(*uwp.PasswordHash), []byte(oldPassword)); err != nil {
			return apperror.Unauthorized("current password is incorrect")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperror.Internal(err)
	}

	return s.queries.SetUserPassword(ctx, userID, string(hash))
}

func (s *AuthService) NeedsSetup(ctx context.Context) (bool, int64, error) {
	count, err := s.queries.CountUsers(ctx)
	if err != nil {
		return false, 0, apperror.Internal(err)
	}
	return count == 0, count, nil
}

func (s *AuthService) GetOrCreateUser(ctx context.Context, email string) (*model.User, error) {
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

	u, err = s.queries.CreateUser(ctx, email, "", isAdmin)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return toModelUser(u), nil
}

func (s *AuthService) DeleteAccount(ctx context.Context, userID uuid.UUID, password string) error {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return apperror.NotFound("user")
		}
		return apperror.Internal(err)
	}

	uwp, err := s.queries.GetUserByEmailWithPassword(ctx, user.Email)
	if err != nil {
		return apperror.Internal(err)
	}

	if uwp.PasswordHash != nil {
		if err := bcrypt.CompareHashAndPassword([]byte(*uwp.PasswordHash), []byte(password)); err != nil {
			return apperror.Unauthorized("password is incorrect")
		}
	}

	if err := s.queries.DeleteUser(ctx, userID); err != nil {
		return apperror.Internal(err)
	}
	return nil
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
		Plan:    ak.Plan,
	}, nil
}

func (s *AuthService) CreateAPIKey(ctx context.Context, userID uuid.UUID, req model.CreateAPIKeyRequest) (*model.APIKey, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if user.Plan != "sponsor" {
		return nil, apperror.Forbidden("API keys require a Sponsor plan")
	}

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

func (s *AuthService) UpdateSettings(ctx context.Context, userID uuid.UUID, settings json.RawMessage) error {
	return s.queries.UpdateUserSettings(ctx, userID, settings)
}

func (s *AuthService) GetSettings(ctx context.Context, userID uuid.UUID) (json.RawMessage, error) {
	return s.queries.GetUserSettings(ctx, userID)
}

func toModelUser(u repository.User) *model.User {
	usr := &model.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		IsAdmin:   u.IsAdmin,
		Plan:      u.Plan,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	if len(u.UserSettings) > 0 {
		raw := json.RawMessage(u.UserSettings)
		usr.UserSettings = &raw
	}
	return usr
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
