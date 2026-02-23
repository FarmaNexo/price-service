# INFRASTRUCTURE.md - Price Service

Documentación de la infraestructura del Price Service.

## Base de Datos

### PostgreSQL
- **Database**: `price_db`
- **Schema**: `price`
- **Puerto**: 5432

### Tablas

#### `price.price_history`
Registra el histórico de precios por producto y farmacia.

| Columna | Tipo | Descripción |
|---|---|---|
| id | UUID PK | Identificador único |
| product_id | UUID NOT NULL | ID del producto (referencia a Catalog) |
| pharmacy_id | UUID NOT NULL | ID de la farmacia (referencia a Pharmacy) |
| pharmacy_name | VARCHAR(255) | Nombre de la farmacia |
| product_name | VARCHAR(500) | Nombre del producto |
| price | DECIMAL(10,2) | Precio actual |
| previous_price | DECIMAL(10,2) | Precio anterior |
| currency | VARCHAR(3) | Moneda (default: PEN) |
| source | VARCHAR(50) | Fuente del precio (inventory_update, manual) |
| recorded_at | TIMESTAMPTZ | Fecha del registro |
| created_at | TIMESTAMPTZ | Fecha de creación |

**Índices**: product_id, pharmacy_id, (product_id, pharmacy_id), recorded_at DESC, (product_id, recorded_at DESC)

#### `price.price_alerts`
Alertas de precio configuradas por usuarios.

| Columna | Tipo | Descripción |
|---|---|---|
| id | UUID PK | Identificador único |
| user_id | UUID NOT NULL | ID del usuario |
| product_id | UUID NOT NULL | ID del producto |
| product_name | VARCHAR(500) | Nombre del producto |
| target_price | DECIMAL(10,2) | Precio objetivo |
| current_price | DECIMAL(10,2) | Precio actual |
| is_active | BOOLEAN | Si la alerta está activa |
| is_triggered | BOOLEAN | Si se ha disparado |
| triggered_at | TIMESTAMPTZ | Cuándo se disparó |
| triggered_pharmacy_id | UUID | Farmacia que disparó |
| triggered_pharmacy_name | VARCHAR(255) | Nombre de la farmacia |
| triggered_price | DECIMAL(10,2) | Precio que disparó |
| created_at | TIMESTAMPTZ | Fecha de creación |
| updated_at | TIMESTAMPTZ | Fecha de actualización |

**Constraint UNIQUE**: (user_id, product_id)

#### `price.generic_brand_comparisons`
Cache de comparaciones genérico vs marca.

| Columna | Tipo | Descripción |
|---|---|---|
| id | UUID PK | Identificador único |
| product_id | UUID NOT NULL | ID del producto |
| product_name | VARCHAR(500) | Nombre del producto |
| active_ingredient | VARCHAR(255) | Ingrediente activo |
| is_generic | BOOLEAN | Si es genérico |
| related_product_id | UUID NOT NULL | ID del producto relacionado |
| related_product_name | VARCHAR(500) | Nombre del producto relacionado |
| related_is_generic | BOOLEAN | Si el relacionado es genérico |
| avg_price_product | DECIMAL(10,2) | Precio promedio del producto |
| avg_price_related | DECIMAL(10,2) | Precio promedio del relacionado |
| savings_percentage | DECIMAL(5,2) | Porcentaje de ahorro |
| last_calculated_at | TIMESTAMPTZ | Última vez calculado |

**Constraint UNIQUE**: (product_id, related_product_id)
**Constraint CHECK**: product_id != related_product_id

## Redis

### Claves de Caché

| Patrón | TTL | Descripción |
|---|---|---|
| `cache:price:compare:{productId}` | 15 min | Comparación de precios |
| `cache:price:history:{productId}:{pharmacyId}:{limit}` | 30 min | Historial de precios |
| `cache:price:alerts:{userId}` | 10 min | Alertas del usuario |
| `cache:price:generic-vs-brand:{productId}` | 1 hora | Genérico vs marca |
| `cache:price:stats:{productId}` | 30 min | Estadísticas |

### Invalidación de Caché
- Al registrar un precio: se invalidan compare, history y stats del producto
- Al crear/eliminar alerta: se invalida alerts del usuario
- Al actualizar comparación: se invalida generic-vs-brand del producto

## AWS SQS

### Colas

| Cola | Dirección | Descripción |
|---|---|---|
| `farmanexo-pharmacy-events` | Consumer | Eventos de inventario del Pharmacy Service |
| `farmanexo-price-events` | Publisher | Eventos de precios publicados |

### Eventos Consumidos

#### `INVENTORY_UPDATED`
```json
{
  "event_type": "INVENTORY_UPDATED",
  "pharmacy_id": "uuid",
  "product_id": "uuid",
  "stock": 100,
  "price": 5.50,
  "timestamp": "2026-02-23T12:00:00Z",
  "metadata": {
    "source": "pharmacy-service",
    "version": "1.0",
    "pharmacy_name": "Farmacia Central",
    "product_name": "Paracetamol 500mg"
  }
}
```

### Eventos Publicados

#### `PRICE_ALERT_TRIGGERED`
```json
{
  "event_type": "PRICE_ALERT_TRIGGERED",
  "product_id": "uuid",
  "pharmacy_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2026-02-23T12:00:00Z",
  "metadata": {
    "source": "price-service",
    "version": "1.0",
    "alert_id": "uuid",
    "target_price": "5.50",
    "triggered_price": "4.99"
  }
}
```

## Comunicación Inter-Servicio

### HTTP Clients

| Servicio | Puerto | Timeout | Endpoints Usados |
|---|---|---|---|
| Pharmacy Service | 4004 | 3s | GET /api/v1/pharmacies/product/{id}/inventory, GET /api/v1/pharmacies/{id} |
| Catalog Service | 4003 | 3s | GET /api/v1/products/{id}, POST /api/v1/products/search |

### Degradación Graceful
- Si Pharmacy Service no responde → retorna lista vacía de precios
- Si Catalog Service no responde → retorna error en búsqueda por ingrediente

## SQS Consumer

### Configuración

| Parámetro | Default | Descripción |
|---|---|---|
| `consumer.enabled` | true | Habilitar/deshabilitar consumer |
| `consumer.poll_interval` | 5s | Intervalo de polling |
| `consumer.max_messages` | 10 | Máximo de mensajes por poll |
| `consumer.visibility_timeout` | 30 | Timeout de visibilidad (segundos) |

### Flujo del Consumer
1. Recibe mensaje `INVENTORY_UPDATED` de la cola
2. Busca precio anterior del producto/farmacia
3. Registra nuevo `PriceHistory` con precio anterior
4. Busca alertas activas para el producto
5. Si precio <= target → dispara alerta, publica evento
6. Si precio > target → actualiza precio actual de la alerta
7. Elimina mensaje de la cola
