package service

import (
	"errors"
	"github.com/KumKeeHyun/gin-realworld/internal/core/domain"
	"github.com/KumKeeHyun/gin-realworld/internal/core/ports"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type commentService struct {
	commentRepo ports.CommentRepository
	articleRepo ports.ArticleRepository
	userRepo    ports.UserRepository
	logger      *zap.SugaredLogger
}

func NewCommentService(
	commentRepo ports.CommentRepository,
	articleRepo ports.ArticleRepository,
	userRepo ports.UserRepository,
	logger *zap.Logger) ports.CommentService {
	return &commentService{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
		userRepo:    userRepo,
		logger:      logger.Sugar().Named("commentService"),
	}
}

func (s commentService) WithTx(tx *gorm.DB) ports.CommentService {
	s.commentRepo = s.commentRepo.WithTx(tx)
	s.userRepo = s.userRepo.WithTx(tx)
	return s
}

func (s commentService) Create(authorID uint, slug string, body string) (domain.CommentView, error) {
	author, err := s.userRepo.FindByID(authorID)
	if err != nil {
		s.logger.Errorw("failed to find user", "err", err)
		return domain.CommentView{}, ports.ErrInternal
	}
	article, err := s.articleRepo.FindBySlug(slug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.CommentView{}, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find article", "err", err)
		return domain.CommentView{}, ports.ErrInternal
	}

	saved, err := s.commentRepo.Save(domain.Comment{
		Body:      body,
		ArticleID: article.ID,
		Author: domain.Author{
			ID:       author.ID,
			Username: author.Username,
			Bio:      author.Bio,
			Image:    author.Image,
		},
	})
	if err != nil {
		s.logger.Errorw("failed to create comment", "err", err)
		return domain.CommentView{}, ports.ErrInternal
	}
	return domain.NewCommentView(saved, false), nil
}

func (s commentService) GetFromArticle(readerID uint, slug string) ([]domain.CommentView, error) {
	_, err := s.articleRepo.FindBySlug(slug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ports.ErrResourceNotFound
	} else if err != nil {
		s.logger.Errorw("failed to find article", "err", err)
		return nil, ports.ErrInternal
	}

	comments, err := s.commentRepo.FindFromArticle(slug)
	if err != nil {
		s.logger.Errorw("failed to find comments", "err", err)
		return nil, ports.ErrInternal
	}

	authorIDs := lo.Map(comments, func(comment domain.Comment, index int) uint { return comment.Author.ID })
	follows, err := s.userRepo.FindFollows(readerID, authorIDs)
	if err != nil {
		s.logger.Errorw("failed to find follows", "err", err)
		return nil, ports.ErrInternal
	}

	return zipToCommentView(comments, follows), nil
}

func zipToCommentView(comments []domain.Comment, follows []domain.Follow) []domain.CommentView {
	followsMap := lo.KeyBy(follows, func(follow domain.Follow) uint { return follow.FollowingID })

	return lo.Map(comments, func(comment domain.Comment, index int) domain.CommentView {
		_, follow := followsMap[comment.Author.ID]
		return domain.NewCommentView(comment, follow)
	})
}

func (s commentService) Delete(authorID, commentID uint) error {
	err := s.commentRepo.Delete(commentID, authorID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Infow("illegal request to delete non-owned comment", "user-id", authorID, "err", err)
		return ports.ErrNonOwnedContent
	} else if err != nil {
		s.logger.Errorw("failed to delete comment", "err", err)
		return ports.ErrInternal
	}
	return nil
}
