// internal/application/handlers/create_price_alert_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/price-service/internal/application/commands"
	"github.com/farmanexo/price-service/internal/domain/entities"
	"github.com/farmanexo/price-service/internal/domain/repositories"
	"github.com/farmanexo/price-service/internal/domain/services"
	"github.com/farmanexo/price-service/internal/presentation/dto/responses"
	"github.com/farmanexo/price-service/internal/shared/common"
	"github.com/farmanexo/price-service/internal/shared/constants"
	"github.com/farmanexo/price-service/pkg/mediator"
	"go.uber.org/zap"
)

const maxAlertsPerUser = 20

// CreatePriceAlertHandler maneja la creación de alertas de precio
type CreatePriceAlertHandler struct {
	alertRepo    repositories.PriceAlertRepository
	cacheService services.CacheService
	logger       *zap.Logger
}

func NewCreatePriceAlertHandler(
	alertRepo repositories.PriceAlertRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *CreatePriceAlertHandler {
	return &CreatePriceAlertHandler{
		alertRepo:    alertRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (h *CreatePriceAlertHandler) Handle(ctx context.Context, cmd commands.CreatePriceAlertCommand) (*common.ApiResponse[responses.PriceAlertResponse], error) {
	// Verificar que no exista ya una alerta para este usuario y producto
	existing, _ := h.alertRepo.FindByUserAndProduct(ctx, cmd.UserID, cmd.ProductID)
	if existing != nil {
		return common.ConflictResponse[responses.PriceAlertResponse](
			constants.CodeAlertAlreadyExists,
			constants.GetDescription(constants.CodeAlertAlreadyExists),
		), nil
	}

	// Verificar límite de alertas
	count, err := h.alertRepo.CountActiveByUser(ctx, cmd.UserID)
	if err != nil {
		h.logger.Error("Error contando alertas activas", zap.Error(err))
		return common.InternalServerErrorResponse[responses.PriceAlertResponse]("Error verificando límite de alertas"), nil
	}
	if count >= maxAlertsPerUser {
		return common.BadRequestResponse[responses.PriceAlertResponse](
			constants.CodeAlertLimitReached,
			constants.GetDescription(constants.CodeAlertLimitReached),
		), nil
	}

	alert := &entities.PriceAlert{
		UserID:      cmd.UserID,
		ProductID:   cmd.ProductID,
		ProductName: cmd.ProductName,
		TargetPrice: cmd.TargetPrice,
		IsActive:    true,
		IsTriggered: false,
	}

	if err := h.alertRepo.Create(ctx, alert); err != nil {
		h.logger.Error("Error creando alerta de precio", zap.Error(err))
		return common.InternalServerErrorResponse[responses.PriceAlertResponse]("Error creando alerta de precio"), nil
	}

	// Invalidar caché de alertas del usuario
	h.cacheService.Delete(ctx, "cache:price:alerts:"+cmd.UserID)

	h.logger.Info("Alerta de precio creada",
		zap.String("alert_id", alert.ID),
		zap.String("user_id", cmd.UserID),
		zap.String("product_id", cmd.ProductID),
		zap.Float64("target_price", cmd.TargetPrice),
	)

	return common.CreatedResponse(responses.ToPriceAlertResponse(*alert)), nil
}

// Compile-time interface check
var _ mediator.RequestHandler[commands.CreatePriceAlertCommand, responses.PriceAlertResponse] = (*CreatePriceAlertHandler)(nil)
