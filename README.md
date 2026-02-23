# FarmaNexo Price Service

Servicio de comparación y seguimiento de precios farmacéuticos para FarmaNexo.

## Descripción

El Price Service es un microservicio Go que permite:
- **Comparar precios** de productos farmacéuticos entre diferentes farmacias
- **Consultar historial** de precios de un producto
- **Crear alertas** de precio (notifica cuando el precio baja al objetivo)
- **Comparar genérico vs marca** para productos con mismo ingrediente activo
- **Estadísticas de precios** (admin) con precio mínimo, máximo y promedio
- **Consumir eventos** de inventario del Pharmacy Service para registrar cambios de precio

## Arquitectura

- **Patrón**: Clean Architecture + CQRS + MediatR
- **Framework**: Chi v5 HTTP Router
- **ORM**: GORM
- **Base de datos**: PostgreSQL (database: `price_db`, schema: `price`)
- **Caché**: Redis
- **Mensajería**: AWS SQS (Publisher + Consumer)
- **Autenticación**: JWT (validación compartida con Auth Service)

## Requisitos

- Go 1.22+
- PostgreSQL 15+
- Redis 7+
- LocalStack (local) o AWS (cloud)

## Instalación

```bash
# Instalar dependencias
make install

# Ejecutar migraciones
make migrate-up

# Ejecutar en modo local
make dev
```

## Configuración

El servicio usa archivos YAML en `configs/`:

| Archivo | Descripción |
|---|---|
| `config.local.yaml` | Desarrollo local |
| `config.development.yaml` | Ambiente de desarrollo |
| `config.qa.yaml` | Quality Assurance |
| `config.uat.yaml` | User Acceptance Testing |
| `config.production.yaml` | Producción |

### Variables de Entorno

| Variable | Descripción |
|---|---|
| `ENV` | Ambiente (local, development, qa, uat, production) |
| `DB_HOST` | Host de PostgreSQL |
| `DB_USER` | Usuario de PostgreSQL |
| `DB_PASSWORD` | Contraseña de PostgreSQL |
| `JWT_SECRET` | Secret compartido para validar JWT |
| `REDIS_HOST` | Host de Redis |
| `REDIS_PASSWORD` | Contraseña de Redis |
| `AWS_REGION` | Región de AWS |
| `SQS_PRICE_EVENTS_URL` | URL de la cola SQS de eventos de precios |
| `SQS_PHARMACY_EVENTS_URL` | URL de la cola SQS de eventos de farmacia |
| `PHARMACY_SERVICE_URL` | URL del Pharmacy Service |
| `CATALOG_SERVICE_URL` | URL del Catalog Service |

## API Endpoints

### Públicos

| Método | Endpoint | Descripción |
|---|---|---|
| POST | `/api/v1/price/compare` | Comparar precios de un producto |
| GET | `/api/v1/price/compare/generic-vs-brand/{id}` | Comparación genérico vs marca |
| GET | `/api/v1/price/history/{id}` | Historial de precios |

### Autenticados (JWT)

| Método | Endpoint | Descripción |
|---|---|---|
| POST | `/api/v1/price/alerts` | Crear alerta de precio |
| GET | `/api/v1/price/alerts` | Listar mis alertas |
| DELETE | `/api/v1/price/alerts/{id}` | Eliminar alerta |

### Admin (JWT + role admin)

| Método | Endpoint | Descripción |
|---|---|---|
| GET | `/api/v1/price/stats/{id}` | Estadísticas de precios |
| POST | `/api/v1/price/record` | Registrar precio manualmente |

### Utilidad

| Método | Endpoint | Descripción |
|---|---|---|
| GET | `/health` | Health check |
| GET | `/swagger/*` | Documentación Swagger |

## Base de Datos

### Tablas

- `price.price_history` — Historial de precios por producto y farmacia
- `price.price_alerts` — Alertas de precio configuradas por usuarios
- `price.generic_brand_comparisons` — Comparaciones genérico vs marca

## Eventos

### Consumidos
- `INVENTORY_UPDATED` (de Pharmacy Service) → Registra precio, verifica alertas

### Publicados
- `PRICE_ALERT_TRIGGERED` → farmanexo-price-events
- `PRICE_COMPARISON_CREATED` → farmanexo-price-events
- `PRICE_RECORDED` → farmanexo-price-events

## Caché Redis

| Clave | TTL | Descripción |
|---|---|---|
| `cache:price:compare:{productId}` | 15 min | Comparación de precios |
| `cache:price:history:{productId}:*` | 30 min | Historial de precios |
| `cache:price:alerts:{userId}` | 10 min | Alertas del usuario |
| `cache:price:generic-vs-brand:{productId}` | 1 hora | Genérico vs marca |
| `cache:price:stats:{productId}` | 30 min | Estadísticas |

## Estructura del Proyecto

```
price-service/
├── cmd/server/main.go              # Entry point
├── internal/
│   ├── application/
│   │   ├── commands/               # 4 commands (CQRS)
│   │   ├── queries/                # 4 queries (CQRS)
│   │   ├── handlers/               # 8 handlers
│   │   ├── validators/             # 2 validators
│   │   ├── preprocessors/          # Input sanitization
│   │   └── postprocessors/         # Audit logging
│   ├── domain/
│   │   ├── entities/               # PriceHistory, PriceAlert, GenericBrandComparison
│   │   ├── repositories/           # 3 repository interfaces
│   │   ├── services/               # CacheService, EventPublisher, PharmacyClient, CatalogClient
│   │   └── events/                 # PriceEvent, InventoryUpdatedEvent
│   ├── infrastructure/
│   │   ├── persistence/postgres/   # 3 GORM repositories
│   │   ├── cache/                  # Redis client + cache service
│   │   ├── messaging/              # SQS publisher + consumer
│   │   ├── clients/                # PharmacyClient, CatalogClient HTTP
│   │   └── security/               # JWT validation
│   ├── presentation/
│   │   ├── http/controllers/       # PriceController
│   │   ├── http/middlewares/       # Auth, CorrelationID
│   │   ├── http/routes/            # Chi routes
│   │   └── dto/                    # Requests + Responses
│   └── shared/
│       ├── common/                 # ApiResponse, ResponseExtensions
│       └── constants/              # MessageCodes, HTTPStatus
├── pkg/
│   ├── config/                     # Configuration
│   └── mediator/                   # MediatR implementation
├── migrations/                     # SQL migrations
├── configs/                        # YAML configs (5 environments)
├── docs/                           # Swagger docs (auto-generated)
├── Makefile
├── Dockerfile
├── CLAUDE.md
└── README.md
```

## Comandos

```bash
make help              # Ver comandos disponibles
make install           # Instalar dependencias
make build             # Compilar binario
make dev               # Ejecutar en modo local
make run               # Ejecutar con swagger
make test              # Ejecutar tests
make lint              # Linter
make swagger           # Generar Swagger docs
make migrate-up        # Aplicar migraciones
make migrate-down      # Revertir migración
make docker-build      # Build Docker image
make docker-run        # Run en Docker
```
