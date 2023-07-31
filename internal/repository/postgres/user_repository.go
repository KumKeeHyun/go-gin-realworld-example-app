package postgres

import (
	"database/sql"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) ports.UserRepository {
	return userRepository{
		db: db,
	}
}

func (r userRepository) WithTx(tx *gorm.DB) ports.UserRepository {
	if tx == nil {
		return r
	}
	r.db = tx
	return r
}

func (r userRepository) Save(user domain.User) (domain.User, error) {
	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"email",
			"username",
			"password",
			"bio",
			"image",
		}),
	}).Create(&user).Error
	return user, err
}

func (r userRepository) FindByID(id uint) (domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	return user, err
}

func (r userRepository) FindByEmail(email string) (domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return user, err
}

func (r userRepository) FindByUsername(username string) (domain.User, error) {
	var user domain.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return user, err
}

func (r userRepository) FindByEmailOrUsername(email, username string) (domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).Or("username = ?", username).First(&user).Error
	return user, err
}

func (r userRepository) FindProfile(curUserID, profileUserID uint) (domain.Profile, error) {
	result := struct {
		ID        uint
		Username  string
		Bio       string
		Image     sql.NullString
		FollowCnt int64
	}{}
	err := r.db.Model(&domain.User{}).
		Select("users.id, users.username, users.bio, users.image, (?) as follow_cnt",
			r.db.Model(&domain.Follow{}).
				Where("follower_id = ?", curUserID).
				Where("following_id = ?", profileUserID).
				Select("count(id)")).
		Where("users.id = ?", profileUserID).
		Scan(&result).Error
	return domain.Profile{
		ID:        result.ID,
		Username:  result.Username,
		Bio:       result.Bio,
		Image:     result.Image,
		Following: result.FollowCnt != 0,
	}, err
}

func (r userRepository) CreateFollow(followerID, followingID uint) (domain.Follow, error) {
	follow := domain.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	return follow, r.db.Create(&follow).Error
}

func (r userRepository) FindFollow(followerID, followingID uint) (domain.Follow, error) {
	var follow domain.Follow
	return follow, r.db.Where("follower_id = ?", followerID).
		Where("following_id = ?", followingID).
		First(&follow).Error
}

func (r userRepository) FindFollows(followerID uint, followingIDs []uint) ([]domain.Follow, error) {
	var follows []domain.Follow
	return follows, r.db.Where("follower_id = ?", followerID).
		Where("following_id IN ?", followingIDs).
		Find(&follows).Error
}

func (r userRepository) DeleteFollow(followerID, followingID uint) error {
	return r.db.
		Where("follower_id = ?", followerID).
		Where("following_id = ?", followingID).
		Delete(&domain.Follow{}).Error
}
