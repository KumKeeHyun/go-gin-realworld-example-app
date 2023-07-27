package service

import (
	"errors"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type profileService struct {
	userRepo ports.UserRepository
	logger   *zap.SugaredLogger
}

func NewProfileService(
	userRepo ports.UserRepository,
	logger *zap.Logger) ports.ProfileService {
	return profileService{
		userRepo: userRepo,
		logger:   logger.Sugar().Named("profileService"),
	}
}

func (s profileService) WithTx(tx *gorm.DB) ports.ProfileService {
	s.userRepo = s.userRepo.WithTx(tx)
	return s
}

func (s profileService) Find(curUserID uint, profileUsername string) (domain.Profile, error) {
	profileUser, err := s.userRepo.FindByUsername(profileUsername)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Profile{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find user by username", "username", profileUsername, "err", err)
		return domain.Profile{}, ports.ErrInternal
	}

	profile, err := s.userRepo.FindProfile(curUserID, profileUser.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Profile{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find profile", "err", err)
		return domain.Profile{}, ports.ErrInternal
	}
	return profile, err
}

func (s profileService) Follow(curUserID uint, followingName string) (domain.Profile, error) {
	following, err := s.userRepo.FindByUsername(followingName)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Profile{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find user by username", "username", followingName, "err", err)
		return domain.Profile{}, ports.ErrInternal
	}

	if curUserID == following.ID {
		s.logger.Infow("illegal request to follow oneself", "user-id", curUserID, "err", err)
		return domain.Profile{}, ports.ErrSelfFollowing
	}

	_, err = s.userRepo.CreateFollow(curUserID, following.ID)
	if err != nil {
		s.logger.Errorw("failed to create follow", "followerID", curUserID, "followingID", following.ID)
		return domain.Profile{}, ports.ErrInternal
	}
	return domain.Profile{
		ID:        following.ID,
		Username:  following.Username,
		Bio:       following.Bio,
		Image:     following.Image,
		Following: true,
	}, nil
}

func (s profileService) Unfollow(curUserID uint, followingName string) (domain.Profile, error) {
	following, err := s.userRepo.FindByUsername(followingName)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Profile{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find user by username", "username", followingName, "err", err)
		return domain.Profile{}, ports.ErrInternal
	}

	err = s.userRepo.DeleteFollow(curUserID, following.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Errorw("failed to delete follow", "followerID", curUserID, "followingID", following.ID)
		return domain.Profile{}, ports.ErrInternal
	}
	return domain.Profile{
		ID:        following.ID,
		Username:  following.Username,
		Bio:       following.Bio,
		Image:     following.Image,
		Following: false,
	}, nil
}
