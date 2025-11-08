package entity

import "github.com/yatori-dev/yatori-go-core/que-core/qentity"

// 选择题转换成标准question
func (q *ChoiceQue) TurnStandardQuestion() qentity.Question {
	question := qentity.Question{
		Type:    q.Type.String(),
		Content: q.Text,
	}
	candidateSelects := []string{} //待选
	resSelect := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N"}
	for _, option := range resSelect {
		if q.Options[option] == "" {
			break
		}
		candidateSelects = append(candidateSelects, q.Options[option])
	}
	question.Options = candidateSelects
	question.Answers = []string{}
	return question
}

// 判断题
func (q *JudgeQue) TurnStandardQuestion() qentity.Question {
	question := qentity.Question{
		Type:    q.Type.String(),
		Content: q.Text,
	}
	candidateSelects := []string{} //待选
	resSelect := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N"}
	for _, option := range resSelect {
		if q.Options[option] == "" {
			break
		}
		candidateSelects = append(candidateSelects, q.Options[option])
	}
	question.Options = candidateSelects
	question.Answers = []string{}
	return question
}

// 填空题
func (q *FillQue) TurnStandardQuestion() qentity.Question {
	question := qentity.Question{
		Type:    q.Type.String(),
		Content: q.Text,
	}
	candidateSelects := []string{}
	for i := 0; i < len(q.OpFromAnswer); i++ {
		candidateSelects = append(candidateSelects, "")
	}
	question.Options = candidateSelects
	return question
}

// 简答题
func (q *ShortQue) TurnStandardQuestion() qentity.Question {
	question := qentity.Question{
		Type:    q.Type.String(),
		Content: q.Text,
		Options: []string{""},
	}
	return question
}

// 名词解释
func (q *TermExplanationQue) TurnStandardQuestion() qentity.Question {
	question := qentity.Question{
		Type:    q.Type.String(),
		Content: q.Text,
		Options: []string{""},
	}
	return question
}

// 论述题
func (q *EssayQue) TurnStandardQuestion() qentity.Question {
	question := qentity.Question{
		Type:    q.Type.String(),
		Content: q.Text,
		Options: []string{""},
	}
	return question
}

// 连线题
func (q *MatchingQue) TurnStandardQuestion() qentity.Question {
	question := qentity.Question{
		Type:    q.Type.String(),
		Content: q.Text,
	}
	candidateSelects := []string{}

	for _, option := range q.Options {
		candidateSelects = append(candidateSelects, "[1]"+option)
	}
	for _, option := range q.Selects {
		candidateSelects = append(candidateSelects, "[2]"+option)
	}
	question.Options = candidateSelects
	return question
}

// 其他题
func (q *OtherQue) TurnStandardQuestion() qentity.Question {
	question := qentity.Question{
		Type:    q.Type.String(),
		Content: q.Text,
		Options: []string{""},
	}
	return question
}
