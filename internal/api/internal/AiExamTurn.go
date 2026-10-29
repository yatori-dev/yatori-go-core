package internal

import (
	"github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
	que_core "github.com/yatori-dev/yatori-go-core/v3/que-core/aiq"
)

type AiExamTurnInterface interface {
	AIProblemMessage(testPaperTitle string, topic xuexitong.ExamTurn) que_core.AIChatMessages
}
