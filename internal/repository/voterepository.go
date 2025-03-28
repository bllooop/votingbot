package repository

import (
	"fmt"

	"github.com/bllooop/votingbot/internal/domain"
	logger "github.com/bllooop/votingbot/pkg/logging"
	"github.com/google/uuid"
	"github.com/tarantool/go-tarantool/v2"
)

type PollsTarantool struct {
	db *tarantool.Connection
}

func NewPollsTarantool(db *tarantool.Connection) *PollsTarantool {
	return &PollsTarantool{
		db: db,
	}
}

func (r *PollsTarantool) CreateDB(question string, options []string, creatorId string) (string, []string, error) {
	pollID := uuid.New().String()
	votes := make([]int, len(options))
	data, err := r.db.Do(
		tarantool.NewInsertRequest("polls").Tuple([]interface{}{pollID, question, options, creatorId, votes, "active"})).Get()
	if err != nil {
		return "", nil, err
	}
	if len(data) == 0 {
		return "", nil, fmt.Errorf("ошибка добавления голосования")
	}
	logger.Log.Debug().Any("poll_id", pollID).Msg("Создано голосование ID:")
	return pollID, options, nil
}

func (r *PollsTarantool) CastDB(pollID string, option string) error {
	resp, err := r.db.Do(
		tarantool.NewSelectRequest("polls").
			Iterator(tarantool.IterEq).
			Key([]interface{}{pollID}),
	).Get()
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return fmt.Errorf("голосование с ID %s не найдено", pollID)
	}

	pollData, ok := resp[0].([]interface{})
	if !ok || len(pollData) < 5 {
		return fmt.Errorf("некорректный формат данных")
	}

	optionsInterface, ok1 := pollData[2].([]interface{})
	if !ok1 {
		return fmt.Errorf("некорректный формат данных вариантов ответа")
	}

	var options []string
	for _, opt := range optionsInterface {
		strOpt, ok := opt.(string)
		if !ok {
			return fmt.Errorf("некорректный формат данных вариантов ответа")
		}
		options = append(options, strOpt)
	}

	votesInterface, ok2 := pollData[4].([]interface{})
	if !ok2 {
		return fmt.Errorf("некорректный формат данных голосоов")
	}

	var votes []int
	for _, v := range votesInterface {
		switch v := v.(type) {
		case int8:
			votes = append(votes, int(v))
		default:
			return fmt.Errorf("некорректный формат данных счета голосов: %T", v)
		}
	}

	if len(votes) == 0 {
		votes = make([]int, len(options))
	}

	if len(options) != len(votes) {
		logger.Log.Error().Msgf("Несовпадение вариантов ответа (%d) и голосов (%d) для голосования %s", len(options), len(votes), pollID)
		return fmt.Errorf("несовпадение данных вариантов ответа и голосов")
	}

	var optionIndex int = -1
	for i, opt := range options {
		if opt == option {
			optionIndex = i
			break
		}
	}
	if optionIndex == -1 {
		return fmt.Errorf("вариант ответа %s не найден в голосовании", option)
	}

	votes[optionIndex]++

	_, err = r.db.Do(
		tarantool.NewUpdateRequest("polls").
			Key([]interface{}{pollID}).
			Operations(tarantool.NewOperations().Assign(4, votes)),
	).Get()
	if err != nil {
		return err
	}

	logger.Log.Debug().Any("poll_id", pollID).Any("option", option).Msg("Голос отдан успешно")
	return nil
}
func (r *PollsTarantool) GetRes(pollID string) (domain.Results, error) {
	data, err := r.db.Do(
		tarantool.NewSelectRequest("polls").
			Limit(1).
			Iterator(tarantool.IterEq).
			Key([]interface{}{pollID}),
	).Get()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("голосование с ID %s не найдено", pollID)
	}

	var results domain.Results
	for _, rawRow := range data {

		row, ok := rawRow.([]interface{})
		if !ok || len(row) < 6 {
			return nil, fmt.Errorf("неожиданный формат данных: %v", rawRow)
		}

		pollID, _ := row[0].(string)
		question, _ := row[1].(string)
		optionsRaw, okOptions := row[2].([]interface{})
		creatorID, _ := row[3].(string)
		votesRaw, okVotes := row[4].([]interface{})
		active, _ := row[5].(string)

		logger.Log.Debug().Msgf("Значения: id=%s, question=%s, options=%v, creator_id=%s, votes=%v, active=%s",
			pollID, question, optionsRaw, creatorID, votesRaw, active)

		var options []string
		if okOptions {
			for _, opt := range optionsRaw {
				if strOpt, ok := opt.(string); ok {
					options = append(options, strOpt)
				}
			}
		}

		var votes []int
		if okVotes {
			for _, v := range votesRaw {
				switch v := v.(type) {
				case int8:
					votes = append(votes, int(v))
				default:
					logger.Log.Error().Msgf("Некорректный тип голосов: %T, значение: %v", v, v)
					return nil, fmt.Errorf("некорректный тип голосов: %T", v)
				}
			}
		}

		if len(votes) < len(options) {
			for len(votes) < len(options) {
				votes = append(votes, 0)
			}
		}

		for i, option := range options {
			results = append(results, domain.Result{
				Question: question,
				Option:   option,
				Count:    votes[i],
			})
		}
	}

	logger.Log.Debug().Any("data", results).Msg("Получены данные о голосовании")
	return results, nil
}

func (r *PollsTarantool) CloseDB(pollID string, creatorId string) error {
	pollData, err := r.getPollByID(pollID)
	if err != nil {
		return err
	}
	creatorID, ok := pollData[3].(string)
	if !ok || creatorID != creatorId {
		return fmt.Errorf("только создатель может закрыть голосование")
	}

	data, err := r.db.Do(
		tarantool.NewUpdateRequest("polls").
			Key([]interface{}{pollID}).
			Operations(tarantool.NewOperations().Assign(5, "Closed")),
	).Get()
	if err != nil {
		return err
	}
	logger.Log.Debug().Any("data", data).Msg("Закрыто голосование")
	return nil
}
func (r *PollsTarantool) DeleteDB(pollID string, creatorId string) error {
	pollData, err := r.getPollByID(pollID)
	if err != nil {
		return err
	}

	creatorID, ok := pollData[3].(string)
	if !ok || creatorID != creatorId {
		return fmt.Errorf("только создатель может удалить голосование")
	}

	_, err = r.db.Do(
		tarantool.NewDeleteRequest("polls").
			Key([]interface{}{pollID}),
	).Get()
	if err != nil {
		return err
	}

	logger.Log.Debug().Msg("Голосование удалено")
	return nil
}

func (r *PollsTarantool) getPollByID(pollID string) ([]interface{}, error) {
	resp, err := r.db.Do(
		tarantool.NewSelectRequest("polls").
			Limit(1).
			Iterator(tarantool.IterEq).
			Key([]interface{}{pollID}),
	).Get()
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, fmt.Errorf("голосование %s не найдено", pollID)
	}

	pollData, ok := resp[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("некорректный формат данных")
	}

	return pollData, nil
}
