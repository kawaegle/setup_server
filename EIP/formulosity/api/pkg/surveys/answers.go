package surveys

import (
	"encoding/json"
	"errors"

	"github.com/plutov/formulosity/api/pkg/services"
	"github.com/plutov/formulosity/api/pkg/types"
)

// returns 2 errors: general and error details
func SubmitAnswer(svc services.Services, session *types.SurveySession, survey *types.Survey, question *types.Question, req []byte, file *types.File) (error, error) {
	logCtx := svc.Logger.With("session_uuid", session.UUID)
	logCtx.Info("submitting answer")

	answer, err := question.GetAnswerType()
	if err != nil {
		return err, nil
	}

	switch a := answer.(type) {
	case *types.FileAnswer:
		if file != nil {
			a.FileSize = file.Size
			a.FileFormat = file.Format

			if err := answer.Validate(*question); err != nil {
				return errors.New("invalid answer"), err
			}

			filePath, err := svc.FileStorage.SaveFile(file)
			if err != nil {
				return errors.New("unable to save file"), nil
			}
			a.AnswerValue = filePath
		} else {
			return errors.New("file is required for this question type"), nil
		}
	default:
		if err := json.Unmarshal(req, &answer); err != nil {
			return errors.New("invalid request format"), nil
		}

		if err := answer.Validate(*question); err != nil {
			return errors.New("invalid answer"), err
		}
	}

	if err := svc.Storage.UpsertSurveyQuestionAnswer(session.UUID, question.UUID, answer); err != nil {
		msg := "unable to insert answer"
		logCtx.Error(msg, "err", err)
		return errors.New(msg), nil
	}

	logCtx.Info("answer submitted")

	// mark session as completed if there are no more unanswered questions
	isCompleted := isSessionCompleted(survey, session, question)

	if isCompleted {
		session.Status = types.SurveySessionStatus_Completed
		if err := svc.Storage.UpdateSurveySessionStatus(session.UUID, session.Status); err != nil {
			msg := "unable to update session status"
			logCtx.Error(msg, "err", err)
			return nil, errors.New(msg)
		}

		logCtx.Info("session completed")
	}

	return nil, nil
}

func isSessionCompleted(survey *types.Survey, session *types.SurveySession, question *types.Question) bool {
	if session.Status == types.SurveySessionStatus_Completed {
		return true
	}

	for _, q := range survey.Config.Questions.Questions {
		hasAnswer := q.UUID == question.UUID
		for _, a := range session.QuestionAnswers {
			if q.UUID == a.QuestionUUID {
				hasAnswer = true
				break
			}
		}

		if !hasAnswer {
			return false
		}
	}

	return true
}
