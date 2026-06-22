package repository

import (
	"context"

	"go-service-starter/domain"

	"gorm.io/gorm"
)

type AnalysisHistoryRepo struct {
	db *gorm.DB
}

func NewAnalysisHistoryRepo(db *gorm.DB) domain.AnalysisHistoryRepository {
	return &AnalysisHistoryRepo{db: db}
}

func (r *AnalysisHistoryRepo) Create(ctx context.Context, h *domain.AnalysisHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *AnalysisHistoryRepo) ListByUserID(ctx context.Context, userID uint, analysisType string, limit, offset int) ([]domain.AnalysisHistory, int64, error) {
	var list []domain.AnalysisHistory
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AnalysisHistory{}).Where("user_id = ?", userID)
	if analysisType != "" {
		query = query.Where("analysis_type = ?", analysisType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *AnalysisHistoryRepo) GetByID(ctx context.Context, id uint) (*domain.AnalysisHistory, error) {
	var h domain.AnalysisHistory
	if err := r.db.WithContext(ctx).First(&h, id).Error; err != nil {
		return nil, err
	}
	return &h, nil
}
