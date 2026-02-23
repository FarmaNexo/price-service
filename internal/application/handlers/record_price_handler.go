// internal/application/handlers/record_price_handler.go
package handlers

import (
	"context"
	"time"

	"github.com/farmanexo/price-service/internal/application/commands"
	"github.com/farmanexo/price-service/internal/domain/entities"
	"github.com/farmanexo/price-service/internal/domain/repositories"
	"github.com/farmanexo/price-service/internal/domain/services"
	"github.com/farmanexo/price-service/internal/presentation/dto/responses"
	"github.com/farmanexo/price-service/internal/shared/common"
	"github.com/farmanexo/price-service/pkg/mediator"
	"go.uber.org/zap"
)

// RecordPriceHandler maneja el registro manual de precios (admin)
type RecordPriceHandler struct {
	priceHistoryRepo repositories.PriceHistoryRepository
	cacheService     services.CacheService
	logger           *zap.Logger
}

func NewRecordPriceHandler(
	priceHistoryRepo repositories.PriceHistoryRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *RecordPriceHandler {
	return &RecordPriceHandler{
		priceHistoryRepo: priceHistoryRepo,
		cacheService:     cacheService,
		logger:           logger,
	}
}

func (h *RecordPriceHandler) Handle(ctx context.Context, cmd commands.RecordPriceCommand) (*common.ApiResponse[responses.PriceHistoryItem], error) {
	// Buscar precio anterior
	var previousPrice *float64
	latestPrice, err := h.priceHistoryRepo.FindLatestPrice(ctx, cmd.ProductID, cmd.PharmacyID)
	if err == nil && latestPrice != nil {
		previousPrice = &latestPrice.Price
	}

	source := cmd.Source
	if source == "" {
		source = "manual"
	}

	history := &entities.PriceHistory{
		ProductID:     cmd.ProductID,
		PharmacyID:    cmd.PharmacyID,
		PharmacyName:  cmd.PharmacyName,
		ProductName:   cmd.ProductName,
		Price:         cmd.Price,
		PreviousPrice: previousPrice,
		Currency:      "PEN",
		Source:        source,
		RecordedAt:    time.Now(),
	}

	if err := h.priceHistoryRepo.Create(ctx, history); err != nil {
		h.logger.Error("Error registrando precio", zap.Error(err))
		return common.InternalServerErrorResponse[responses.PriceHistoryItem]("Error registrando precio"), nil
	}

	// Invalidar cachés relacionados
	h.cacheService.DeleteByPattern(ctx, "cache:price:compare:"+cmd.ProductID+"*")
	h.cacheService.DeleteByPattern(ctx, "cache:price:history:"+cmd.ProductID+"*")
	h.cacheService.DeleteByPattern(ctx, "cache:price:stats:"+cmd.ProductID+"*")

	h.logger.Info("Precio registrado manualmente",
		zap.String("product_id", cmd.ProductID),
		zap.String("pharmacy_id", cmd.PharmacyID),
		zap.Float64("price", cmd.Price),
	)

	return common.CreatedResponse(responses.ToPriceHistoryItem(*history)), nil
}

// Compile-time interface check
var _ mediator.RequestHandler[commands.RecordPriceCommand, responses.PriceHistoryItem] = (*RecordPriceHandler)(nil)
