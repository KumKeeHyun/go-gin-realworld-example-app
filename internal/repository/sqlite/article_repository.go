package sqlite

import (
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type articleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ports.ArticleRepository {
	return articleRepository{
		db: db,
	}
}

func (r articleRepository) WithTx(tx *gorm.DB) ports.ArticleRepository {
	if tx == nil {
		return r
	}
	r.db = tx
	return r
}

func (r articleRepository) Save(article domain.Article) (domain.Article, error) {
	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"slug",
			"title",
			"description",
			"body",
		}),
	}).Create(&article).Error
	return article, err
}

func (r articleRepository) FindBySlug(slug string) (domain.Article, error) {
	var article domain.Article
	return article, r.db.Where("slug = ?", slug).First(&article).Error
}

func (r articleRepository) FindBySearchConditions(cond ports.ArticleSearchConditions) ([]domain.Article, error) {
	var ids []uint
	tx := r.db.Model(&domain.Article{})
	if cond.Tag != nil {
		tx = tx.Where("tags LIKE ?", "%"+*cond.Tag+"%")
	}
	if cond.Author != nil {
		tx = tx.Where("author_username = ?", *cond.Author)
	}
	if cond.Favorited != nil {
		tx = tx.Where("id IN (?)", r.db.Model(&domain.Favorite{}).
			Where("user_id = (?)", r.db.Model(&domain.User{}).
				Where("username = ?", *cond.Favorited).
				Select("id")).
			Select("article_id"))
	}
	err := tx.Limit(cond.Limit).Offset(cond.Offset).Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	} else if len(ids) == 0 {
		return nil, nil
	}

	var articles []domain.Article
	return articles, r.db.Where("id in ?", ids).Find(&articles).Error
}

func (r articleRepository) FindFeed(userID uint, pageable ports.Pageable) ([]domain.Article, error) {
	var ids []uint
	tx := r.db.Model(&domain.Article{})
	tx.Where("author_id IN (?)", r.db.Model(&domain.Follow{}).
		Where("follower_id = ?", userID).
		Select("following_id"))
	err := tx.Limit(pageable.Limit).Offset(pageable.Offset).Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	} else if len(ids) == 0 {
		return nil, nil
	}

	var articles []domain.Article
	return articles, r.db.Where("id in ?", ids).Find(&articles).Error
}

func (r articleRepository) DeleteBySlug(slug string) error {
	return r.db.
		Where("slug = ?", slug).
		Delete(&domain.Article{}).Error
}

func (r articleRepository) UpdateAuthorInfo(user domain.User) error {
	return r.db.Model(&domain.Article{}).
		Where("author_id = ?", user.ID).Updates(
		map[string]any{
			"author_username": user.Username,
			"author_bio":      user.Bio,
			"author_image":    user.Image,
		},
	).Error
}

func (r articleRepository) CreateFavorite(userID, articleID uint) (domain.Favorite, error) {
	favorite := domain.Favorite{
		UserID:    userID,
		ArticleID: articleID,
	}
	return favorite, r.db.Create(&favorite).Error
}

func (r articleRepository) FindFavorite(userID uint, articleID uint) (domain.Favorite, error) {
	var favorite domain.Favorite
	return favorite, r.db.Where("user_id = ?", userID).
		Where("article_id = ?", articleID).
		First(&favorite).Error
}

func (r articleRepository) FindFavorites(userID uint, articleIDs []uint) ([]domain.Favorite, error) {
	var favorites []domain.Favorite
	return favorites, r.db.Where("user_id = ?", userID).
		Where("article_id IN ?", articleIDs).
		Find(&favorites).Error
}

func (r articleRepository) DeleteFavorite(userID, articleID uint) error {
	return r.db.
		Where("user_id = ?", userID).
		Where("article_id = ?", articleID).
		Delete(&domain.Favorite{}).Error
}

// FindTags only local test purpose
func (r articleRepository) FindTags() ([]string, error) {
	var tags []string
	err := r.db.Raw(`
		WITH RECURSIVE split(value, str) AS (
			SELECT null, rtrim(ltrim(tags, '{'), '}') || ',' FROM articles
			UNION ALL
			SELECT
				substr(str, 0, instr(str, ',')),
				substr(str, instr(str, ',')+1)
			FROM split WHERE str!=''
		) SELECT DISTINCT trim(value, '"') as tag FROM split WHERE value is not NULL
	`).Pluck("tag", &tags).Error
	return tags, err
}
