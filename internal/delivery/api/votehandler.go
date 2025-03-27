package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bllooop/votingbot/internal/domain"
	logger "github.com/bllooop/votingbot/pkg/logging"
	"github.com/gin-gonic/gin"
)

func (h *Handler) VoteHandler(c *gin.Context) {
	logger.Log.Info().Msg("Получен запрос в бота")
	if c.Request.Method != http.MethodPost {
		logger.Log.Error().Msg("Требуется запрос POST")
		newErrorResponse(c, http.StatusBadRequest, "Требуется запрос POST")
		return
	}

	var req domain.MattermostRequest
	if err := c.ShouldBind(&req); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "Недействительные данные в запросе")
		return
	}

	args := strings.Fields(req.Text)
	if len(args) == 0 {
		newErrorResponse(c, http.StatusBadRequest, "Ошибка: нужно указать тип команды")
		return
	}
	switch args[0] {
	case "create":
		h.createPoll(c, req, args[1:])
	case "cast":
		h.castVote(c, req, args[1:])
	case "results":
		h.getResults(c, req, args[1:])
	case "close":
		h.closePoll(c, req, args[1:])
	case "delete":
		h.deletePoll(c, req, args[1:])
	default:
		c.JSON(http.StatusOK, gin.H{"response_type": "ephemeral", "text": "Неизвестная команда"})
	}

}

func (h *Handler) createPoll(c *gin.Context, req domain.MattermostRequest, args []string) {
	if len(args) < 2 {
		newErrorResponse(c, http.StatusBadRequest, "Ошибка: нужно указать вопрос и хотя бы два варианта ответа")
		return
	}
	question := args[0]
	var options []string
	questionEnded := false
	for i := 1; i < len(args); i++ {
		if !questionEnded && strings.HasSuffix(args[i], "?") {
			question += " " + args[i]
			questionEnded = true
		} else if questionEnded {
			options = append(options, args[i])
		} else {
			question += " " + args[i]
		}
	}
	optionsText := req.Text
	optionsList := strings.Split(optionsText, ",")
	options = nil
	for i := range optionsList {
		optionsList[i] = strings.TrimSpace(optionsList[i])
	}

	if len(optionsList) < 2 {
		newErrorResponse(c, http.StatusBadRequest, "Ошибка: необходимо указать хотя бы два варианта ответа")
		return
	}

	options = append(options, optionsList...)

	logger.Log.Info().Msgf("Получен запрос на создание голосования с данными %s, %s", question, options)
	pollID, err := h.Usecases.Polls.CreateDB(question, options, req.UserID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	responseText := fmt.Sprintf("Голосование создано! ID: %s", pollID)
	c.JSON(http.StatusOK, domain.MattermostResponse{
		ResponseType: "in_channel",
		Text:         responseText,
	})
}

func (h *Handler) castVote(c *gin.Context, req domain.MattermostRequest, args []string) {
	if len(args) < 2 {
		newErrorResponse(c, http.StatusBadRequest, "Ошибка: укажите ID голосования и вариант ответа")
		return
	}
	pollID := args[0]
	option := strings.Join(args[1:], " ")
	logger.Log.Info().Msgf("Получен запрос на выбор варианта %s в голосовании %s", option, pollID)
	err := h.Usecases.Polls.CastDB(pollID, option)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	responseText := fmt.Sprintf("<@%s> проголосовал за %s в голосовании %s", req.UserID, option, pollID)
	c.JSON(http.StatusOK, domain.MattermostResponse{
		ResponseType: "in_channel",
		Text:         responseText,
	})
}

func (h *Handler) getResults(c *gin.Context, req domain.MattermostRequest, args []string) {
	if len(args) < 1 {
		newErrorResponse(c, http.StatusBadRequest, "Ошибка: укажите ID голосования")
		return
	}
	pollID := args[0]
	logger.Log.Info().Msgf("Получен запрос на данные о голосовании %s", pollID)
	results, err := h.Usecases.Polls.GetRes(pollID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	var resultText, responseText string
	if len(results) == 0 {
		responseText = fmt.Sprintf("Результаты голосования %s: нет данных", pollID)
	} else {
		question := results[0].Question
		for _, res := range results {
			resultText += fmt.Sprintf("%s: %d голосов\n", res.Option, res.Count)
		}
		responseText = fmt.Sprintf("Результаты голосования  %s, %s:\n%s", pollID, question, resultText)
	}

	c.JSON(http.StatusOK, domain.MattermostResponse{
		ResponseType: "in_channel",
		Text:         responseText,
	})
}

func (h *Handler) closePoll(c *gin.Context, req domain.MattermostRequest, args []string) {
	if len(args) < 1 {
		newErrorResponse(c, http.StatusBadRequest, "Ошибка: укажите ID голосования")
		return
	}
	pollID := args[0]
	logger.Log.Info().Msgf("Получен запрос на закрытие голосования %s", pollID)
	err := h.Usecases.Polls.CloseDB(pollID, req.UserID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	responseText := fmt.Sprintf("Голосование %s закрыто", pollID)

	c.JSON(http.StatusOK, domain.MattermostResponse{
		ResponseType: "in_channel",
		Text:         responseText,
	})
}

func (h *Handler) deletePoll(c *gin.Context, req domain.MattermostRequest, args []string) {
	if len(args) < 1 {
		newErrorResponse(c, http.StatusBadRequest, "Ошибка: укажите ID голосования")
		return
	}
	pollID := args[0]
	logger.Log.Info().Msgf("Получен запрос на удаление голосования %s", pollID)
	err := h.Usecases.Polls.DeleteDB(pollID, req.UserID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	responseText := fmt.Sprintf("Голосование %s удаленео", pollID)

	c.JSON(http.StatusOK, domain.MattermostResponse{
		ResponseType: "in_channel",
		Text:         responseText,
	})
}
