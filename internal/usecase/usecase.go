package usecase

import (
	"github.com/bllooop/votingbot/internal/domain"
	"github.com/bllooop/votingbot/internal/repository"
)

type Polls interface {
	CreateDB(question string, options []string, creatorId string) (string, []string, error)
	CastDB(pollID string, option string) error
	GetRes(pollID string) (domain.Results, error)
	CloseDB(pollID string, creatorId string) error
	DeleteDB(pollID string, creatorId string) error
}
type Usecase struct {
	Polls
}

func NewUsecase(repo *repository.Repository) *Usecase {
	return &Usecase{
		Polls: NewPollsUsecase(repo),
	}
}
