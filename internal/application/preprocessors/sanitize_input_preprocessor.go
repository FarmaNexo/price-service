// internal/application/preprocessors/sanitize_input_preprocessor.go
package preprocessors

import (
	"context"
	"strings"

	"github.com/farmanexo/price-service/internal/application/commands"
	"go.uber.org/zap"
)

// SanitizeInputPreProcessor limpia y normaliza los inputs
type SanitizeInputPreProcessor struct {
	logger *zap.Logger
}

func NewSanitizeInputPreProcessor(logger *zap.Logger) *SanitizeInputPreProcessor {
	return &SanitizeInputPreProcessor{logger: logger}
}

func (p *SanitizeInputPreProcessor) Process(ctx context.Context, request interface{}) error {
	switch cmd := request.(type) {
	case *commands.CreatePriceAlertCommand:
		cmd.ProductID = strings.TrimSpace(cmd.ProductID)
		cmd.ProductName = strings.TrimSpace(cmd.ProductName)
		p.logger.Debug("Input sanitizado", zap.String("command", "CreatePriceAlertCommand"))

	case *commands.ComparePricesCommand:
		cmd.ProductID = strings.TrimSpace(cmd.ProductID)
		p.logger.Debug("Input sanitizado", zap.String("command", "ComparePricesCommand"))

	case *commands.RecordPriceCommand:
		cmd.ProductID = strings.TrimSpace(cmd.ProductID)
		cmd.PharmacyID = strings.TrimSpace(cmd.PharmacyID)
		cmd.PharmacyName = strings.TrimSpace(cmd.PharmacyName)
		cmd.ProductName = strings.TrimSpace(cmd.ProductName)
		cmd.Source = strings.TrimSpace(cmd.Source)
		p.logger.Debug("Input sanitizado", zap.String("command", "RecordPriceCommand"))
	}

	return nil
}
