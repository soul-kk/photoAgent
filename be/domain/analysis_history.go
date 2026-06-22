package domain

import (
	"context"
	"time"
)

type AnalysisType string

const (
	AnalysisTypeAnalyze      AnalysisType = "analyze"
	AnalysisTypeScore        AnalysisType = "score"
	AnalysisTypeShootAdvice  AnalysisType = "shoot_advice"
	AnalysisTypeCompare      AnalysisType = "compare"
	AnalysisTypeToneStyle    AnalysisType = "tone_style"
	AnalysisTypeRecognize    AnalysisType = "recognize"
)

type AnalysisHistory struct {
	ID           uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint         `gorm:"index;not null" json:"user_id"`
	AnalysisType AnalysisType `gorm:"size:32;not null" json:"analysis_type"`
	InputPrompt  string       `gorm:"type:text" json:"input_prompt"`
	FocusDimension string     `gorm:"size:64" json:"focus_dimension"`
	ResultJSON   string       `gorm:"type:text" json:"result_json"`
	Score        *float64     `gorm:"type:decimal(5,2)" json:"score,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

type AnalysisHistoryRepository interface {
	Create(ctx context.Context, h *AnalysisHistory) error
	ListByUserID(ctx context.Context, userID uint, analysisType string, limit, offset int) ([]AnalysisHistory, int64, error)
	GetByID(ctx context.Context, id uint) (*AnalysisHistory, error)
}
