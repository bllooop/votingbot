package usecase

import (
	"github.com/bllooop/votingbot/internal/domain"
	"github.com/bllooop/votingbot/internal/repository"
)

type PollsUsecase struct {
	repo repository.Polls
}

func NewPollsUsecase(repo *repository.Repository) *PollsUsecase {
	return &PollsUsecase{
		repo: repo,
	}
}
func (s *PollsUsecase) CreateDB(question string, options []string, creatorId string) (string, error) {
	return s.repo.CreateDB(question, options, creatorId)
}
func (s *PollsUsecase) CastDB(pollID string, option string) error {
	return s.repo.CastDB(pollID, option)
}
func (s *PollsUsecase) GetRes(pollID string) (domain.Results, error) {
	return s.repo.GetRes(pollID)
}
func (s *PollsUsecase) CloseDB(pollID string, creatorId string) error {
	return s.repo.CloseDB(pollID, creatorId)
}
func (s *PollsUsecase) DeleteDB(pollID string, creatorId string) error {
	return s.repo.DeleteDB(pollID, creatorId)
}
