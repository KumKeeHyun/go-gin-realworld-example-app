package sqlite

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"gorm.io/gorm"
)

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) ports.CommentRepository {
	return &commentRepository{db: db}
}

func (r commentRepository) WithTx(tx *gorm.DB) ports.CommentRepository {
	if tx == nil {
		return r
	}
	r.db = tx
	return r
}

func (r commentRepository) Save(comment domain.Comment) (domain.Comment, error) {
	return comment, r.db.Save(&comment).Error
}

func (r commentRepository) FindFromArticle(slug string) ([]domain.Comment, error) {
	var ids []int
	err := r.db.Model(&domain.Comment{}).
		Where("article_id = (?)", r.db.Model(&domain.Article{}).
			Where("slug = ?", slug).
			Select("id")).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}

	var comments []domain.Comment
	return comments, r.db.Where("id IN ?", ids).Find(&comments).Error
}

func (r commentRepository) Delete(id, authorID uint) error {
	tx := r.db.Where("id = ?", id).
		Where("author_id = ?", authorID).
		Delete(&domain.Comment{})
	if tx.Error != nil {
		return tx.Error
	} else if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
