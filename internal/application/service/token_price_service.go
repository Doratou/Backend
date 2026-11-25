package service

import (
	"context"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_repo "github.com/Doratou/Backend/internal/domain/repository/DB"
)

type tokenPriceService struct {
	repo domain_repo.TokenPriceRepository
}

// NewTokenPriceService creates a new instance of TokenPriceService.
func NewTokenPriceService(repo domain_repo.TokenPriceRepository) adapter.TokenPriceService {
	return &tokenPriceService{repo: repo}
}

func (s *tokenPriceService) UpdatePrice(ctx context.Context, adminUserID int64, req dto.UpdateTokenPriceRequest) (*dto.TokenPriceResponse, error) {
	if adminUserID <= 0 || req.PriceToman <= 0 {
		return nil, exception.ErrBadRequest
	}

	oldPrice, updated, err := s.repo.UpdatePrice(ctx, req.PriceToman)
	if err != nil {
		return nil, err
	}

	return &dto.TokenPriceResponse{
		OldPriceToman: oldPrice,
		NewPriceToman: updated.PriceToman,
		UpdatedAt:     updated.UpdatedAt,
	}, nil
}
