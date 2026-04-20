// backend/internal/game/types/helpers.go

package types

import (
	"time"

	"github.com/google/uuid"
)

func NewMCQAnswer(questionID uuid.UUID, slug string, timeSpent time.Duration) Answer {
	return Answer{
		QuestionID: questionID,
		AnswerType: "mcq",
		AnswerData: slug,
		TimeSpent:  timeSpent,
		AnsweredAt: time.Now(),
	}
}

func NewTextAnswer(questionID uuid.UUID, text string, timeSpent time.Duration) Answer {
	return Answer{
		QuestionID: questionID,
		AnswerType: "text",
		AnswerData: text,
		TimeSpent:  timeSpent,
		AnsweredAt: time.Now(),
	}
}

func NewBooleanAnswer(questionID uuid.UUID, value bool, timeSpent time.Duration) Answer {
	return Answer{
		QuestionID: questionID,
		AnswerType: "boolean",
		AnswerData: value,
		TimeSpent:  timeSpent,
		AnsweredAt: time.Now(),
	}
}


func NewMultipleAnswer(questionID uuid.UUID, slugs []string, timeSpent time.Duration) Answer {
	return Answer{
		QuestionID: questionID,
		AnswerType: "multiple",
		AnswerData: slugs,
		TimeSpent:  timeSpent,
		AnsweredAt: time.Now(),
	}
}
