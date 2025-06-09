package controllers

import (
	"errors"
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/plutov/formulosity/api/pkg/http/response"
	surveyspkg "github.com/plutov/formulosity/api/pkg/surveys"
	"github.com/plutov/formulosity/api/pkg/types"
)

func (h *Handler) getSurvey(c echo.Context) error {
	survey, err := h.getLaunchedSurvey(c)
	if err != nil {
		return response.NotFound(c, err.Error())
	}

	return response.Ok(c, survey)
}

func (h *Handler) getSurveyCSS(c echo.Context) error {
	survey, err := h.getLaunchedSurvey(c)
	if err != nil {
		return response.NotFound(c, err.Error())
	}

	// serve css
	if survey.Config.Theme == types.Theme_Custom {
		filePath := fmt.Sprintf("%s/%s/theme.css", os.Getenv("SURVEYS_DIR"), survey.Name)
		return c.File(filePath)
	}

	return response.Ok(c, "ok")
}

func (h *Handler) getLaunchedSurvey(c echo.Context) (*types.Survey, error) {
	urlSlug := c.Param("url_slug")
	res, err := surveyspkg.GetSurvey(h.Services, urlSlug)
	if err != nil {
		return nil, errors.New("survey not found")
	}

	if res.DeliveryStatus != types.SurveyDeliveryStatus_Launched {
		return nil, errors.New("survey is stopped")
	}

	if res.Config == nil {
		return nil, errors.New("invalid survey configuration")
	}

	return res, nil
}

type updateSurveyReq struct {
	DeliveryStatus types.SurveyDeliveryStatus `json:"delivery_status"`
}

func (h *Handler) getSurveys(c echo.Context) error {
	surveys, err := h.Storage.GetSurveys()
	if err != nil {
		h.Logger.Error("failed to get surveys", "err", err)
		return response.InternalErrorDefaultMsg(c)
	}

	for i, s := range surveys {
		surveys[i].URL = fmt.Sprintf("/survey/%s", s.URLSlug)
	}

	return response.Ok(c, surveys)
}

func (h *Handler) updateSurvey(c echo.Context) error {
	survey := c.Get("survey").(types.Survey)

	req := new(updateSurveyReq)
	if err := c.Bind(req); err != nil {
		return response.BadRequestDefaultMessage(c)
	}
	if req.DeliveryStatus != types.SurveyDeliveryStatus_Launched && req.DeliveryStatus != types.SurveyDeliveryStatus_Stopped {
		return response.BadRequest(c, "invalid delivery status")
	}

	updateSurvey := &survey
	updateSurvey.DeliveryStatus = req.DeliveryStatus

	err := surveyspkg.UpdateSurvey(h.Services, updateSurvey)
	if err != nil {
		return response.InternalErrorDefaultMsg(c)
	}

	return response.Ok(c, survey)
}
