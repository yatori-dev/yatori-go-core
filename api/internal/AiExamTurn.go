package internal

import (
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/utils"
)

type AiExamTurnInterface interface {
	AIProblemMessage(testPaperTitle string, topic entity.ExamTurn) utils.AIChatMessages
}
