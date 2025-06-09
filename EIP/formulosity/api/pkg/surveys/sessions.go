package surveys

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/plutov/formulosity/api/pkg/services"
	"github.com/plutov/formulosity/api/pkg/types"
)

func CreateSurveySession(svc services.Services, survey *types.Survey, ipAddr string) (*types.SurveySession, error) {
	session := &types.SurveySession{
		Status:     types.SurveySessionStatus_InProgress,
		SurveyUUID: survey.UUID,
		IPAddr:     ipAddr,
	}

	logCtx := svc.Logger.With("session", *session)
	logCtx.Info("creating survey session")

	if ipAddr != "" && survey.Config.Security.DuplicateProtection == types.DuplicateProtectionType_Ip {
		if ipAddrSession, _ := svc.Storage.GetSurveySessionByIPAddress(survey.UUID, ipAddr); ipAddrSession != nil {
			msg := "duplicate session for ip address"
			logCtx.Error(msg)
			return nil, errors.New(msg)
		}
	}

	if err := svc.Storage.CreateSurveySession(session); err != nil {
		msg := "unable to create survey session"
		logCtx.Error(msg, "err", err)
		return nil, errors.New(msg)
	}

	logCtx.With("session_uuid", session.UUID).Info("survey session created")

	return session, nil
}

func GetSurveySession(svc services.Services, survey types.Survey, sessionUUID string) (*types.SurveySession, error) {
	logCtx := svc.Logger.With("survey_uuid", survey.UUID, "session_uuid", sessionUUID)
	logCtx.Info("getting survey session")

	session, err := svc.Storage.GetSurveySession(survey.UUID, sessionUUID)
	if err != nil {
		msg := "session not found"
		logCtx.Error(msg, "err", err)
		return nil, errors.New(msg)
	}

	if session == nil {
		return nil, errors.New("session not found")
	}

	answers, err := svc.Storage.GetSurveySessionAnswers(session.UUID)
	if err != nil {
		msg := "unable to get session answers"
		logCtx.Error(msg, "err", err)
	}

	session.QuestionAnswers = answers

	session.QuestionAnswers = convertAnswerBytesToAnswerType(svc, &survey, session.QuestionAnswers)

	return session, nil
}

func convertAnswerBytesToAnswerType(svc services.Services, survey *types.Survey, answers []types.QuestionAnswer) []types.QuestionAnswer {
	for i, a := range answers {
		for _, q := range survey.Config.Questions.Questions {
			if q.UUID == a.QuestionUUID {
				answerType, err := q.GetAnswerType()
				if err != nil {
					svc.Logger.Error("unable to get answer type", "err", err)
				} else {
					if err := json.Unmarshal(a.AnswerBytes, &answerType); err != nil {
						svc.Logger.Error("unable to decode answer", "err", err)
					}
					answers[i].Answer = answerType
				}

				break
			}
		}
	}

	return answers
}

func GetSurveySessions(svc services.Services, survey types.Survey, filter *types.SurveySessionsFilter) ([]types.SurveySession, int, error) {
	logCtx := svc.Logger.With("survey_uuid", survey.UUID)
	logCtx.Info("getting survey sessions")

	sessions, totalCount, err := svc.Storage.GetSurveySessionsWithAnswers(survey.UUID, filter)
	if err != nil {
		msg := "unable to get survey sessions"
		logCtx.Error(msg, "err", err)
		return nil, 0, errors.New(msg)
	}

	for i, s := range sessions {
		sessions[i].QuestionAnswers = convertAnswerBytesToAnswerType(svc, &survey, s.QuestionAnswers)
	}

	pagesCount := totalCount / filter.Limit

	return sessions, pagesCount, nil
}

func CallWebhook(svc services.Services, survey *types.Survey, session *types.SurveySession) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("invalid post data, err: %v", err)
	}

	req, err := http.NewRequest(survey.Config.Webhook.Method, survey.Config.Webhook.URL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("invalid http request, err: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request, err: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			svc.Logger.Error("unable to close body", "err", err)
		}
	}()

	statusCode := resp.StatusCode
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		responseBody = []byte{}
	}

	return svc.Storage.StoreWebhookResponse(int(session.ID), statusCode, string(responseBody))
}
