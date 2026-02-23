// internal/shared/constants/message_codes.go
package constants

// MessageCode contiene todos los códigos de respuesta del sistema
type MessageCode string

const (
	// Success codes
	CodeSuccess        MessageCode = "SUCCESS_001"
	CodeCreatedSuccess MessageCode = "SUCCESS_002"
	CodeUpdatedSuccess MessageCode = "SUCCESS_003"
	CodeDeletedSuccess MessageCode = "SUCCESS_004"

	// Price domain codes
	CodePriceCompared          MessageCode = "PRC_001"
	CodePriceHistoryRetrieved  MessageCode = "PRC_002"
	CodeAlertCreated           MessageCode = "PRC_003"
	CodeAlertsListed           MessageCode = "PRC_004"
	CodeAlertDeleted           MessageCode = "PRC_005"
	CodeGenericVsBrandRetrieved MessageCode = "PRC_006"
	CodePriceStatsRetrieved    MessageCode = "PRC_007"
	CodePriceRecorded          MessageCode = "PRC_008"
	CodeAlertTriggered         MessageCode = "PRC_009"
	CodeComparisonCreated      MessageCode = "PRC_010"

	// Validation errors
	CodeValidationError MessageCode = "VAL_001"
	CodeRequiredField   MessageCode = "VAL_006"
	CodeInvalidPrice    MessageCode = "VAL_007"
	CodeInvalidPage     MessageCode = "VAL_011"

	// Authentication errors
	CodeUnauthorized MessageCode = "AUTH_ERR_001"
	CodeInvalidToken MessageCode = "AUTH_ERR_002"
	CodeTokenExpired MessageCode = "AUTH_ERR_003"
	CodeForbidden    MessageCode = "AUTH_ERR_005"

	// Business errors
	CodeProductNotFound      MessageCode = "BUS_001"
	CodeResourceNotFound     MessageCode = "BUS_003"
	CodeAlertNotFound        MessageCode = "BUS_010"
	CodeAlertAlreadyExists   MessageCode = "BUS_011"
	CodePharmacyNotFound     MessageCode = "BUS_012"
	CodeNoPriceData          MessageCode = "BUS_013"
	CodeAlertLimitReached    MessageCode = "BUS_014"

	// Rate limiting
	CodeRateLimitExceeded MessageCode = "RATE_001"

	// System errors
	CodeInternalError      MessageCode = "SYS_001"
	CodeDatabaseError      MessageCode = "SYS_002"
	CodeServiceUnavailable MessageCode = "SYS_003"
	CodeCacheError         MessageCode = "SYS_005"
	CodeExternalServiceError MessageCode = "SYS_006"
)

// MessageDescription contiene las descripciones predefinidas
var MessageDescription = map[MessageCode]string{
	// Success
	CodeSuccess:        "Operación exitosa",
	CodeCreatedSuccess: "Recurso creado exitosamente",
	CodeUpdatedSuccess: "Recurso actualizado exitosamente",
	CodeDeletedSuccess: "Recurso eliminado exitosamente",

	// Price
	CodePriceCompared:           "Comparación de precios obtenida exitosamente",
	CodePriceHistoryRetrieved:   "Historial de precios obtenido exitosamente",
	CodeAlertCreated:            "Alerta de precio creada exitosamente",
	CodeAlertsListed:            "Alertas de precio listadas exitosamente",
	CodeAlertDeleted:            "Alerta de precio eliminada exitosamente",
	CodeGenericVsBrandRetrieved: "Comparación genérico vs marca obtenida exitosamente",
	CodePriceStatsRetrieved:     "Estadísticas de precios obtenidas exitosamente",
	CodePriceRecorded:           "Precio registrado exitosamente",
	CodeAlertTriggered:          "Alerta de precio disparada",
	CodeComparisonCreated:       "Comparación de precios creada exitosamente",

	// Validation
	CodeValidationError: "Error de validación",
	CodeRequiredField:   "Campo requerido",
	CodeInvalidPrice:    "Precio inválido",
	CodeInvalidPage:     "Paginación inválida",

	// Auth errors
	CodeUnauthorized: "No autorizado",
	CodeInvalidToken: "Token inválido",
	CodeTokenExpired: "Token expirado",
	CodeForbidden:    "No tiene permisos para esta acción",

	// Business
	CodeProductNotFound:    "Producto no encontrado",
	CodeResourceNotFound:   "Recurso no encontrado",
	CodeAlertNotFound:      "Alerta de precio no encontrada",
	CodeAlertAlreadyExists: "Ya existe una alerta para este producto",
	CodePharmacyNotFound:   "Farmacia no encontrada",
	CodeNoPriceData:        "No hay datos de precios disponibles",
	CodeAlertLimitReached:  "Límite máximo de alertas alcanzado",

	// Rate limiting
	CodeRateLimitExceeded: "Demasiadas solicitudes. Intente nuevamente más tarde",

	// System
	CodeInternalError:        "Error interno del servidor",
	CodeDatabaseError:        "Error de base de datos",
	CodeServiceUnavailable:   "Servicio no disponible",
	CodeCacheError:           "Error en servicio de caché",
	CodeExternalServiceError: "Error en servicio externo",
}

// GetDescription retorna la descripción del código
func GetDescription(code MessageCode) string {
	if desc, ok := MessageDescription[code]; ok {
		return desc
	}
	return "Descripción no disponible"
}
