// internal/application/postprocessors/log_audit_postprocessor.go
package postprocessors

import (
	"context"
	"time"

	"github.com/farmanexo/price-service/internal/application/commands"
	"github.com/farmanexo/price-service/internal/application/queries"
	"github.com/farmanexo/price-service/pkg/mediator"
	"go.uber.org/zap"
)

// LogAuditPostProcessor registra eventos de auditoría
type LogAuditPostProcessor struct {
	logger *zap.Logger
}

func NewLogAuditPostProcessor(logger *zap.Logger) *LogAuditPostProcessor {
	return &LogAuditPostProcessor{logger: logger}
}

func (p *LogAuditPostProcessor) Process(ctx context.Context, request interface{}, response interface{}) error {
	userID := p.getUserIDFromContext(ctx)
	correlationID := mediator.GetCorrelationID(ctx)
	isSuccess := p.checkSuccess(response)

	switch request.(type) {
	case commands.ComparePricesCommand, *commands.ComparePricesCommand:
		p.logAudit("PRICES_COMPARED", userID, correlationID, isSuccess)
	case commands.CreatePriceAlertCommand, *commands.CreatePriceAlertCommand:
		p.logAudit("PRICE_ALERT_CREATED", userID, correlationID, isSuccess)
	case commands.DeletePriceAlertCommand, *commands.DeletePriceAlertCommand:
		p.logAudit("PRICE_ALERT_DELETED", userID, correlationID, isSuccess)
	case commands.RecordPriceCommand, *commands.RecordPriceCommand:
		p.logAudit("PRICE_RECORDED", userID, correlationID, isSuccess)
	case queries.GetPriceHistoryQuery, *queries.GetPriceHistoryQuery:
		p.logAudit("PRICE_HISTORY_QUERIED", userID, correlationID, isSuccess)
	case queries.GetGenericVsBrandQuery, *queries.GetGenericVsBrandQuery:
		p.logAudit("GENERIC_VS_BRAND_QUERIED", userID, correlationID, isSuccess)
	case queries.GetPriceStatsQuery, *queries.GetPriceStatsQuery:
		p.logAudit("PRICE_STATS_QUERIED", userID, correlationID, isSuccess)
	default:
		p.logger.Debug("Post-processor: comando sin auditoría configurada")
	}

	return nil
}

func (p *LogAuditPostProcessor) logAudit(eventType, userID, correlationID string, success bool) {
	p.logger.Info("AUDIT",
		zap.String("event_type", eventType),
		zap.Bool("success", success),
		zap.String("correlation_id", correlationID),
		zap.String("user_id", userID),
		zap.Time("timestamp", time.Now()),
	)
}

func (p *LogAuditPostProcessor) checkSuccess(response interface{}) bool {
	if resp, ok := response.(interface{ IsValid() bool }); ok {
		return resp.IsValid()
	}
	return false
}

func (p *LogAuditPostProcessor) getUserIDFromContext(ctx context.Context) string {
	userID, _ := mediator.GetUserID(ctx)
	if userID == "" {
		return "ANONYMOUS"
	}
	return userID
}
