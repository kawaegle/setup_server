package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
)

type QuestionType string

const (
	QuestionType_DropdownSingle   QuestionType = "single-choice"
	QuestionType_DropdownMultiple QuestionType = "multiple-choice"
	QuestionType_ShortText        QuestionType = "short-text"
	QuestionType_LongText         QuestionType = "long-text"
	QuestionType_Date             QuestionType = "date"
	QuestionType_Rating           QuestionType = "rating"
	QuestionType_Ranking          QuestionType = "ranking"
	QuestionType_YesNo            QuestionType = "yes-no"
	QuestionType_Email            QuestionType = "email"
	QuestionType_File             QuestionType = "file"
)

var supportedQuestionTypes = map[QuestionType]bool{
	QuestionType_DropdownSingle:   true,
	QuestionType_DropdownMultiple: true,
	QuestionType_ShortText:        true,
	QuestionType_LongText:         true,
	QuestionType_Date:             true,
	QuestionType_Rating:           true,
	QuestionType_Ranking:          true,
	QuestionType_YesNo:            true,
	QuestionType_Email:            true,
	QuestionType_File:             true,
}

type Questions struct {
	Questions []Question `json:"questions" yaml:"questions"`
}

type Question struct {
	Type                QuestionType        `json:"type" yaml:"type"`
	Label               string              `json:"label" yaml:"label"`
	ID                  string              `json:"id" yaml:"id"`
	Description         string              `json:"description" yaml:"description"`
	Min                 *int                `json:"min,omitempty" yaml:"min,omitempty"`
	Max                 *int                `json:"max,omitempty" yaml:"max,omitempty"`
	OptionsFromVariable *string             `json:"-" yaml:"optionsFromVariable,omitempty"`
	Options             []string            `json:"options,omitempty" yaml:"options,omitempty"`
	UUID                string              `json:"uuid" yaml:"-"`
	Validation          *QuestionValidation `json:"validation,omitempty" yaml:"validation,omitempty"`
}

type QuestionValidation struct {
	Min          *int      `json:"min,omitempty" yaml:"min,omitempty"`
	Max          *int      `json:"max,omitempty" yaml:"max,omitempty"`
	Formats      *[]string `json:"formats,omitempty" yaml:"formats,omitempty"`
	MaxSizeBytes *string   `json:"max_size_bytes,omitempty" yaml:"max_size_bytes,omitempty"`
}

func (s *Questions) Validate() error {
	if len(s.Questions) == 0 {
		return fmt.Errorf("at least one question is required")
	}
	uniqueIDs := make(map[string]bool)
	for _, q := range s.Questions {
		if _, ok := supportedQuestionTypes[q.Type]; !ok {
			return fmt.Errorf("questions[].type is invalid: %s", q.Type)
		}

		if q.Label == "" {
			return fmt.Errorf("questions[].label is required")
		}

		if q.ID != "" {
			if _, ok := uniqueIDs[q.ID]; ok {
				return fmt.Errorf("questions[].id must be unique")
			}
			uniqueIDs[q.ID] = true
		}

		if q.Validation != nil {
			if err := q.Validation.Validate(); err != nil {
				return err
			}
		}

		switch q.Type {
		case QuestionType_DropdownSingle:
		case QuestionType_DropdownMultiple:
		case QuestionType_Ranking:
			if err := q.ValidateOptions(); err != nil {
				return err
			}
		case QuestionType_Rating:
			if err := q.ValidateMinMax(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v QuestionValidation) ValidateFile() error {
	if v.Formats == nil || len(*v.Formats) == 0 {
		return fmt.Errorf("questions[].validation.formats is required")
	}

	if v.MaxSizeBytes != nil {
		if *v.MaxSizeBytes == "" {
			return fmt.Errorf("questions[].validation.maxSizeBytes is empty")
		}
		bytes, err := GetStringMultiplication(*v.MaxSizeBytes)
		if err != nil {
			return fmt.Errorf("questions[].validation.maxSizeBytes is invalid: %w", err)
		}
		if bytes <= 0 {
			return fmt.Errorf("questions[].validation.maxSizeBytes must be greater than or equal to 0")
		}
	} else {
		return fmt.Errorf("questions[].validation.maxSizeBytes is required when questions[].type is file")
	}

	for _, fileType := range *v.Formats {
		if !strings.HasPrefix(fileType, ".") {
			return fmt.Errorf("questions[].validation.formats is invalid: %s", fileType)
		}
	}
	return nil
}

func (q Question) GenerateHash() string {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(q); err != nil {
		return ""
	}

	h := sha256.New()
	h.Write(b.Bytes())
	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

func (q Question) ValidateOptions() error {
	uniqueOptions := make(map[string]bool)
	for _, option := range q.Options {
		if len(option) == 0 {
			return fmt.Errorf("questions[].options must not be empty")
		}
		if _, ok := uniqueOptions[option]; ok {
			return fmt.Errorf("questions[].options must be unique")
		}
		uniqueOptions[option] = true
	}
	if len(uniqueOptions) == 0 {
		return fmt.Errorf("questions[].options must have at least one option")
	}

	return nil
}

func (q Question) ValidateMinMax() error {
	if q.Min == nil || *q.Min == 0 {
		return fmt.Errorf("questions[].min is required")
	}
	if q.Max == nil || *q.Max == 0 {
		return fmt.Errorf("questions[].max is required")
	}
	if *q.Min < 0 {
		return fmt.Errorf("questions[].min must be greater than or equal to 0")
	}
	if *q.Max < 0 {
		return fmt.Errorf("questions[].max must be greater than or equal to 0")
	}
	if *q.Min > *q.Max {
		return fmt.Errorf("questions[].min must be less than or equal to questions[].max")
	}

	return nil
}

func (v QuestionValidation) Validate() error {
	if v.Min != nil && *v.Min < 0 {
		return fmt.Errorf("questions[].validation.min must be greater than or equal to 0")
	}
	if v.Max != nil && *v.Max < 0 {
		return fmt.Errorf("questions[].validation.max must be greater than or equal to 0")
	}
	if v.Min != nil && v.Max != nil && *v.Min > *v.Max {
		return fmt.Errorf("questions[].validation.min must be less than or equal to questions[].validation.max")
	}

	return nil
}

func (q Question) GetAnswerType() (Answer, error) {
	switch q.Type {
	case QuestionType_DropdownSingle:
		return &SingleOptionAnswer{}, nil
	case QuestionType_DropdownMultiple:
		return &MultiOptionsAnswer{}, nil
	case QuestionType_ShortText:
		return &TextAnswer{}, nil
	case QuestionType_LongText:
		return &TextAnswer{}, nil
	case QuestionType_Date:
		return &DateAnswer{}, nil
	case QuestionType_Rating:
		return &NumberAnswer{}, nil
	case QuestionType_Ranking:
		return &MultiOptionsAnswer{}, nil
	case QuestionType_YesNo:
		return &BoolAnswer{}, nil
	case QuestionType_Email:
		return &EmailAnswer{}, nil
	case QuestionType_File:
		return &FileAnswer{}, nil
	default:
		return nil, fmt.Errorf("question type %s is not supported", q.Type)
	}
}

func (q Question) ValidateAnswer(answer interface{}) error {
	return nil
}

func GetStringMultiplication(expression string) (int64, error) {
	parts := strings.Split(expression, "*")

	var result int64 = 1

	for _, part := range parts {
		part = strings.TrimSpace(part)

		num, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return 0, err
		}
		result *= num
	}

	return int64(result), nil
}
