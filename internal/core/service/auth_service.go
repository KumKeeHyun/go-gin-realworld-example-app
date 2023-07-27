package service

import (
	"database/sql"
	"errors"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/KumKeeHyun/gin-realworld/pkg/jwtutil"
	"github.com/KumKeeHyun/gin-realworld/pkg/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type authService struct {
	userRepo    ports.UserRepository
	articleRepo ports.ArticleRepository
	jwtUtil     *jwtutil.JwtUtil
	logger      *zap.SugaredLogger
}

func NewAuthService(
	userRepo ports.UserRepository,
	articleRepo ports.ArticleRepository,
	jwtUtil *jwtutil.JwtUtil,
	logger *zap.Logger) ports.AuthService {
	return authService{
		userRepo:    userRepo,
		articleRepo: articleRepo,
		jwtUtil:     jwtUtil,
		logger:      logger.Sugar().Named("authService"),
	}
}

func (s authService) WithTx(tx *gorm.DB) ports.AuthService {
	s.userRepo = s.userRepo.WithTx(tx)
	s.articleRepo = s.articleRepo.WithTx(tx)
	return s
}

func (s authService) Register(email, username, password string) (domain.User, error) {
	_, err := s.userRepo.FindByEmailOrUsername(email, username)
	if err == nil {
		s.logger.Infow("failed to register user due to duplicated identifier", "email", email, "username", username)
		return domain.User{}, ports.ErrDuplicatedEmailOrUsername
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Errorw("failed to find user by email or username", "err", err)
		return domain.User{}, ports.ErrInternal
	}

	saved, err := s.userRepo.Save(domain.User{
		Email:    email,
		Username: username,
		Password: types.Password{String: password},
	})
	if err != nil {
		s.logger.Errorw("failed to save user", "err", err)
		return domain.User{}, ports.ErrInternal
	}

	saved.Token, err = s.jwtUtil.SignClaims(saved.AccessClaim())
	if err != nil {
		s.logger.Errorw("failed to generate jwt token", "err", err)
		return domain.User{}, ports.ErrInternal
	}
	return saved, nil
}

func (s authService) Login(email, password string) (domain.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find user by email", "email", email, "err", err)
		return domain.User{}, ports.ErrInternal
	}

	if !user.ValidPassword(password) {
		return domain.User{}, ports.ErrInvalidPassword
	}

	user.Token, err = s.jwtUtil.SignClaims(user.AccessClaim())
	if err != nil {
		s.logger.Errorw("failed to generate jwt token", "err", err)
		return domain.User{}, ports.ErrInternal
	}
	return user, nil
}

func (s authService) Update(userID uint, fields ports.UserUpdateFields) (domain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find user by id", "id", userID, "err", err)
		return domain.User{}, ports.ErrInternal
	}

	updatedUser := updateUserFields(user, fields)
	saved, err := s.userRepo.Save(updatedUser)
	if err != nil {
		s.logger.Errorw("failed to save user", "id", userID, "err", err)
		return domain.User{}, ports.ErrInternal
	}

	err = s.articleRepo.UpdateAuthorInfo(saved)
	if err != nil {
		s.logger.Errorw("failed to update author info", "id", userID, "err", err)
		return domain.User{}, ports.ErrInternal
	}
	return saved, nil
}

func updateUserFields(user domain.User, fields ports.UserUpdateFields) domain.User {
	if fields.Email != nil {
		user.Email = *fields.Email
	}
	if fields.Username != nil {
		user.Username = *fields.Username
	}
	if fields.Password != nil {
		user.UpdatePassword(*fields.Password)
	}
	if fields.Bio != nil {
		user.Bio = *fields.Bio
	}
	if fields.Image != nil {
		user.Image = sql.NullString{String: *fields.Email, Valid: true}
	}
	return user
}
