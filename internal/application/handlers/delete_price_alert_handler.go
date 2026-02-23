// internal/application/handlers/delete_price_alert_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/price-service/internal/application/commands"
	"github.com/farmanexo/price-service/internal/domain/repositories"
	"github.com/farmanexo/price-service/internal/domain/services"
	"github.com/farmanexo/price-service/internal/presentation/dto/responses"
	"github.com/farmanexo/price-service/internal/shared/common"
	"github.com/farmanexo/price-service/internal/shared/constants"
	"github.com/farmanexo/price-service/pkg/mediator"
	"go.uber.org/zap"
)

// DeletePriceAlertHandler maneja la eliminación de alertas de precio
type DeletePriceAlertHandler struct {
	alertRepo    repositories.PriceAlertRepository
	cacheService services.CacheService
	logger       *zap.Logger
}

func NewDeletePriceAlertHandler(
	alertRepo repositories.PriceAlertRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *DeletePriceAlertHandler {
	return &DeletePriceAlertHandler{
		alertRepo:    alertRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (h *DeletePriceAlertHandler) Handle(ctx context.Context, cmd commands.DeletePriceAlertCommand) (*common.ApiResponse[responses.DeleteAlertResponse], error) {
	if err := h.alertRepo.Delete(ctx, cmd.AlertID, cmd.UserID); err != nil {
		h.logger.Warn("Alerta no encontrada o no pertenece al usuario",
			zap.String("alert_id", cmd.AlertID),
			zap.String("user_id", cmd.UserID),
		)
		return common.NotFoundResponse[responses.DeleteAlertResponse](
			constants.GetDescription(constants.CodeAlertNotFound),
		), nil
	}

	// Invalidar caché de alertas del usuario
	h.cacheService.Delete(ctx, "cache:price:alerts:"+cmd.UserID)

	h.logger.Info("Alerta de precio eliminada",
		zap.String("alert_id", cmd.AlertID),
		zap.String("user_id", cmd.UserID),
	)

	result := responses.DeleteAlertResponse{
		AlertID: cmd.AlertID,
		Deleted: true,
	}

	resp := common.OkResponse(result)
	resp.AddMessage(constants.CodeAlertDeleted, constants.GetDescription(constants.CodeAlertDeleted))
	return resp, nil
}

// Compile-time interface check
var _ mediator.RequestHandler[commands.DeletePriceAlertCommand, responses.DeleteAlertResponse] = (*DeletePriceAlertHandler)(nil)
