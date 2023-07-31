package service

import (
	"errors"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/KumKeeHyun/gin-realworld/pkg/slugutil"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type articleService struct {
	articleRepo ports.ArticleRepository
	userRepo    ports.UserRepository
	logger      *zap.SugaredLogger
}

func NewArticleService(
	articleRepo ports.ArticleRepository,
	userRepo ports.UserRepository,
	logger *zap.Logger) ports.ArticleService {
	return articleService{
		articleRepo: articleRepo,
		userRepo:    userRepo,
		logger:      logger.Sugar().Named("articleService"),
	}
}

func (s articleService) WithTx(tx *gorm.DB) ports.ArticleService {
	s.articleRepo = s.articleRepo.WithTx(tx)
	s.userRepo = s.userRepo.WithTx(tx)
	return s
}

func (s articleService) Create(authorID uint, title, description, body string, tags []string) (domain.ArticleView, error) {
	author, err := s.userRepo.FindByID(authorID)
	if err != nil {
		s.logger.Errorw("failed to create article", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}
	article := domain.Article{
		Slug:        slugutil.Make(title),
		Title:       title,
		Description: description,
		Body:        body,
		Tags:        tags,
		Author: domain.Author{
			ID:       author.ID,
			Username: author.Username,
			Bio:      author.Bio,
			Image:    author.Image,
		},
	}
	saved, err := s.articleRepo.Save(article)
	if err != nil {
		s.logger.Errorw("failed to create article", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}
	return domain.NewArticleView(saved, false, false), nil
}

func (s articleService) Find(readerID uint, slug string) (domain.ArticleView, error) {
	article, err := s.articleRepo.FindBySlug(slug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ArticleView{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find article", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}

	_, favoriteErr := s.articleRepo.FindFavorite(readerID, article.ID)
	if err != nil && !errors.Is(favoriteErr, gorm.ErrRecordNotFound) {
		s.logger.Errorw("failed to find favorite", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}
	_, followErr := s.userRepo.FindFollow(readerID, article.Author.ID)
	if err != nil && !errors.Is(followErr, gorm.ErrRecordNotFound) {
		s.logger.Errorw("failed to find follow", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}

	return domain.NewArticleView(
		article,
		favoriteErr == nil,
		followErr == nil,
	), nil
}

func (s articleService) ListByConditions(readerID uint, conditions ports.ArticleSearchConditions) ([]domain.ArticleView, error) {
	articles, err := s.articleRepo.FindBySearchConditions(conditions)
	if err != nil {
		s.logger.Errorw("failed to search article", "conditions", conditions, "err", err)
		return nil, ports.ErrInternal
	} else if len(articles) == 0 {
		return nil, nil
	}

	articleIDs := lo.Map(articles, func(article domain.Article, index int) uint { return article.ID })
	favorites, err := s.articleRepo.FindFavorites(readerID, articleIDs)
	if err != nil {
		s.logger.Errorw("failed to find favorites", "err", err)
		return nil, ports.ErrInternal
	}

	authorIDs := lo.Map(articles, func(article domain.Article, index int) uint { return article.Author.ID })
	follows, err := s.userRepo.FindFollows(readerID, authorIDs)
	if err != nil {
		s.logger.Errorw("failed to find follows", "err", err)
		return nil, ports.ErrInternal
	}

	articleViews := zipToArticleView(articles, favorites, follows)
	return articleViews, err
}

func zipToArticleView(articles []domain.Article, favorites []domain.Favorite, follows []domain.Follow) []domain.ArticleView {
	favoritesMap := lo.KeyBy(favorites, func(favorite domain.Favorite) uint { return favorite.ArticleID })
	followsMap := lo.KeyBy(follows, func(follow domain.Follow) uint { return follow.FollowingID })

	return lo.Map(articles, func(article domain.Article, index int) domain.ArticleView {
		_, favorite := favoritesMap[article.ID]
		_, follow := followsMap[article.Author.ID]
		return domain.NewArticleView(article, favorite, follow)
	})
}

func (s articleService) ListFeed(readerID uint, pageable ports.Pageable) ([]domain.ArticleView, error) {
	articles, err := s.articleRepo.FindFeed(readerID, pageable)
	if err != nil {
		s.logger.Errorw("failed to search feed", "err", err)
		return nil, ports.ErrInternal
	} else if len(articles) == 0 {
		return nil, nil
	}

	articleIDs := lo.Map(articles, func(article domain.Article, index int) uint { return article.ID })
	favorites, err := s.articleRepo.FindFavorites(readerID, articleIDs)
	if err != nil {
		s.logger.Errorw("failed to find favorites", "err", err)
		return nil, ports.ErrInternal
	}

	authorIDs := lo.Map(articles, func(article domain.Article, index int) uint { return article.Author.ID })
	follows, err := s.userRepo.FindFollows(readerID, authorIDs)
	if err != nil {
		s.logger.Errorw("failed to find follows", "err", err)
		return nil, ports.ErrInternal
	}

	articleViews := zipToArticleView(articles, favorites, follows)
	return articleViews, err
}

func (s articleService) Update(authorID uint, slug string, fields ports.ArticleUpdateFields) (domain.ArticleView, error) {
	article, err := s.articleRepo.FindBySlug(slug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ArticleView{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find article", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}

	if article.Author.ID != authorID {
		s.logger.Infow("illegal request to update non-owned article", "user-id", authorID, "err", err)
		return domain.ArticleView{}, ports.ErrNonOwnedContent
	}

	updated := updateArticleFields(article, fields)
	updated, err = s.articleRepo.Save(updated)
	if err != nil {
		s.logger.Errorw("failed to update article", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}

	_, favoriteErr := s.articleRepo.FindFavorite(authorID, article.ID)
	if err != nil && !errors.Is(favoriteErr, gorm.ErrRecordNotFound) {
		s.logger.Errorw("failed to find favorite", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}
	return domain.NewArticleView(article, favoriteErr == nil, false), nil
}

func updateArticleFields(article domain.Article, fields ports.ArticleUpdateFields) domain.Article {
	if fields.Title != nil {
		article.Title = *fields.Title
		article.Slug = slugutil.Make(article.Title)
	}
	if fields.Description != nil {
		article.Description = *fields.Description
	}
	if fields.Body != nil {
		article.Body = *fields.Body
	}
	return article
}

func (s articleService) Delete(authorID uint, slug string) error {
	article, err := s.articleRepo.FindBySlug(slug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find article", "err", err)
		return ports.ErrInternal
	}

	if article.Author.ID != authorID {
		s.logger.Infow("illegal request to delete non-owned article", "user-id", authorID, "err", err)
		return ports.ErrNonOwnedContent
	}

	err = s.articleRepo.DeleteBySlug(slug)
	if err != nil {
		s.logger.Errorw("failed to delete article", "err", err)
		return ports.ErrInternal
	}
	return nil
}

func (s articleService) Favorite(userID uint, slug string) (domain.ArticleView, error) {
	article, err := s.articleRepo.FindBySlug(slug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ArticleView{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find article", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}

	_, err = s.articleRepo.CreateFavorite(userID, article.ID)
	if err != nil {
		s.logger.Errorw("failed to create favorite", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}

	_, followErr := s.userRepo.FindFollow(userID, article.Author.ID)
	if followErr != nil && !errors.Is(followErr, gorm.ErrRecordNotFound) {
		s.logger.Errorw("failed to find follow", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}
	return domain.NewArticleView(article, true, followErr == nil), nil
}

func (s articleService) Unfavorite(userID uint, slug string) (domain.ArticleView, error) {
	article, err := s.articleRepo.FindBySlug(slug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ArticleView{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find article", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}

	err = s.articleRepo.DeleteFavorite(userID, article.ID)
	if err != nil {
		s.logger.Errorw("failed to delete favorite", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}

	_, followErr := s.userRepo.FindFollow(userID, article.Author.ID)
	if followErr != nil && !errors.Is(followErr, gorm.ErrRecordNotFound) {
		s.logger.Errorw("failed to find follow", "err", err)
		return domain.ArticleView{}, ports.ErrInternal
	}
	return domain.NewArticleView(article, false, followErr == nil), nil
}

func (s articleService) ListTags() ([]string, error) {
	tags, err := s.articleRepo.FindTags()
	if err != nil {
		s.logger.Errorw("failed to find tags", "err", err)
		return nil, ports.ErrInternal
	}
	return tags, nil
}
