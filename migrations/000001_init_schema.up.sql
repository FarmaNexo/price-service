-- ========================================
-- PRICE SERVICE - Schema inicial
-- ========================================

-- Crear schema
CREATE SCHEMA IF NOT EXISTS price;

-- ========================================
-- TABLA: price_history
-- Registra el histórico de precios por producto y farmacia
-- ========================================
CREATE TABLE IF NOT EXISTS price.price_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    pharmacy_id UUID NOT NULL,
    pharmacy_name VARCHAR(255) NOT NULL DEFAULT '',
    product_name VARCHAR(500) NOT NULL DEFAULT '',
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    previous_price DECIMAL(10,2) CHECK (previous_price >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'PEN',
    source VARCHAR(50) NOT NULL DEFAULT 'inventory_update',
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Índices para price_history
CREATE INDEX idx_price_history_product_id ON price.price_history(product_id);
CREATE INDEX idx_price_history_pharmacy_id ON price.price_history(pharmacy_id);
CREATE INDEX idx_price_history_product_pharmacy ON price.price_history(product_id, pharmacy_id);
CREATE INDEX idx_price_history_recorded_at ON price.price_history(recorded_at DESC);
CREATE INDEX idx_price_history_product_recorded ON price.price_history(product_id, recorded_at DESC);

-- ========================================
-- TABLA: price_alerts
-- Alertas de precio configuradas por usuarios
-- ========================================
CREATE TABLE IF NOT EXISTS price.price_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    product_id UUID NOT NULL,
    product_name VARCHAR(500) NOT NULL DEFAULT '',
    target_price DECIMAL(10,2) NOT NULL CHECK (target_price > 0),
    current_price DECIMAL(10,2) CHECK (current_price >= 0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_triggered BOOLEAN NOT NULL DEFAULT false,
    triggered_at TIMESTAMP WITH TIME ZONE,
    triggered_pharmacy_id UUID,
    triggered_pharmacy_name VARCHAR(255),
    triggered_price DECIMAL(10,2),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_price_alert_user_product UNIQUE (user_id, product_id)
);

-- Índices para price_alerts
CREATE INDEX idx_price_alerts_user_id ON price.price_alerts(user_id);
CREATE INDEX idx_price_alerts_product_id ON price.price_alerts(product_id);
CREATE INDEX idx_price_alerts_active ON price.price_alerts(is_active) WHERE is_active = true;
CREATE INDEX idx_price_alerts_user_active ON price.price_alerts(user_id, is_active) WHERE is_active = true;

-- ========================================
-- TABLA: generic_brand_comparisons
-- Cache de comparaciones genérico vs marca
-- ========================================
CREATE TABLE IF NOT EXISTS price.generic_brand_comparisons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    product_name VARCHAR(500) NOT NULL DEFAULT '',
    active_ingredient VARCHAR(255) NOT NULL,
    is_generic BOOLEAN NOT NULL DEFAULT false,
    related_product_id UUID NOT NULL,
    related_product_name VARCHAR(500) NOT NULL DEFAULT '',
    related_is_generic BOOLEAN NOT NULL DEFAULT false,
    avg_price_product DECIMAL(10,2) NOT NULL DEFAULT 0,
    avg_price_related DECIMAL(10,2) NOT NULL DEFAULT 0,
    savings_percentage DECIMAL(5,2) NOT NULL DEFAULT 0,
    last_calculated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_generic_brand_pair UNIQUE (product_id, related_product_id),
    CONSTRAINT chk_different_products CHECK (product_id != related_product_id)
);

-- Índices para generic_brand_comparisons
CREATE INDEX idx_gbc_product_id ON price.generic_brand_comparisons(product_id);
CREATE INDEX idx_gbc_active_ingredient ON price.generic_brand_comparisons(active_ingredient);
CREATE INDEX idx_gbc_related_product_id ON price.generic_brand_comparisons(related_product_id);
