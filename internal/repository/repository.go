package repository

import (
	"github.com/bllooop/votingbot/internal/domain"
	"github.com/tarantool/go-tarantool/v2"
)

type Polls interface {
	CreateDB(question string, options []string, creatorId string) (string, error)
	CastDB(pollID string, option string) error
	GetRes(pollID string) (domain.Results, error)
	CloseDB(pollID string, creatorId string) error
	DeleteDB(pollID string, creatorId string) error
}

type Repository struct {
	Polls
}

func NewRepository(db *tarantool.Connection) *Repository {
	return &Repository{
		Polls: NewPollsTarantool(db),
	}
}
