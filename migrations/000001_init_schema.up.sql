-- Customers (Kirana Stores)
CREATE TABLE IF NOT EXISTS customers (
    id                      BIGSERIAL PRIMARY KEY,
    name                    VARCHAR(255)    NOT NULL,
    owner_name              VARCHAR(255)    NOT NULL,
    phone                   VARCHAR(15)     NOT NULL UNIQUE,
    email                   VARCHAR(255)    UNIQUE,
    gst_number              VARCHAR(20)     UNIQUE,
    store_type              VARCHAR(50)     NOT NULL DEFAULT 'general',
    credit_limit            DECIMAL(12,2)   NOT NULL DEFAULT 0,
    preferred_delivery_slot VARCHAR(50),
    lat                     DECIMAL(10,7)   NOT NULL CHECK (lat BETWEEN -90 AND 90),
    lng                     DECIMAL(10,7)   NOT NULL CHECK (lng BETWEEN -180 AND 180),
    address_line1           VARCHAR(255),
    city                    VARCHAR(100),
    state                   VARCHAR(100),
    pincode                 VARCHAR(10),
    is_active               BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Sellers
CREATE TABLE IF NOT EXISTS sellers (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(255)    NOT NULL,
    owner_name      VARCHAR(255)    NOT NULL,
    phone           VARCHAR(15)     NOT NULL UNIQUE,
    email           VARCHAR(255)    UNIQUE,
    gst_number      VARCHAR(20)     UNIQUE,
    business_type   VARCHAR(50)     NOT NULL DEFAULT 'wholesaler',
    lat             DECIMAL(10,7)   NOT NULL CHECK (lat BETWEEN -90 AND 90),
    lng             DECIMAL(10,7)   NOT NULL CHECK (lng BETWEEN -180 AND 180),
    address_line1   VARCHAR(255),
    city            VARCHAR(100),
    state           VARCHAR(100),
    pincode         VARCHAR(10),
    rating          DECIMAL(3,2)    NOT NULL DEFAULT 0.0,
    is_verified     BOOLEAN         NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Products
CREATE TABLE IF NOT EXISTS products (
    id                      BIGSERIAL PRIMARY KEY,
    seller_id               BIGINT          NOT NULL REFERENCES sellers(id) ON DELETE CASCADE,
    name                    VARCHAR(255)    NOT NULL,
    description             TEXT,
    sku                     VARCHAR(100)    NOT NULL,
    category                VARCHAR(50)     NOT NULL DEFAULT 'other',
    mrp                     DECIMAL(10,2)   NOT NULL CHECK (mrp > 0),
    selling_price           DECIMAL(10,2)   NOT NULL CHECK (selling_price > 0),
    bulk_price              DECIMAL(10,2),
    actual_weight_kg        DECIMAL(8,3)    NOT NULL CHECK (actual_weight_kg > 0),
    length_cm               DECIMAL(8,2)    NOT NULL CHECK (length_cm > 0),
    width_cm                DECIMAL(8,2)    NOT NULL CHECK (width_cm > 0),
    height_cm               DECIMAL(8,2)    NOT NULL CHECK (height_cm > 0),
    volumetric_weight_kg    DECIMAL(8,3)    GENERATED ALWAYS AS
                                ((length_cm * width_cm * height_cm) / 5000) STORED,
    is_fragile              BOOLEAN         NOT NULL DEFAULT FALSE,
    is_perishable           BOOLEAN         NOT NULL DEFAULT FALSE,
    stock_quantity          INT             NOT NULL DEFAULT 0,
    min_order_quantity      INT             NOT NULL DEFAULT 1,
    is_active               BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    UNIQUE(seller_id, sku)
);

-- Warehouses
CREATE TABLE IF NOT EXISTS warehouses (
    id                      BIGSERIAL PRIMARY KEY,
    name                    VARCHAR(255)    NOT NULL UNIQUE,
    code                    VARCHAR(50)     NOT NULL UNIQUE,
    lat                     DECIMAL(10,7)   NOT NULL CHECK (lat BETWEEN -90 AND 90),
    lng                     DECIMAL(10,7)   NOT NULL CHECK (lng BETWEEN -180 AND 180),
    address_line1           VARCHAR(255),
    city                    VARCHAR(100),
    state                   VARCHAR(100),
    pincode                 VARCHAR(10),
    contact_person          VARCHAR(255),
    contact_phone           VARCHAR(15),
    max_capacity_sqft       INT,
    current_load_percent    DECIMAL(5,2)    DEFAULT 0 CHECK (current_load_percent BETWEEN 0 AND 100),
    is_active               BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Transport Rate Config
CREATE TABLE IF NOT EXISTS transport_rates (
    id                  BIGSERIAL PRIMARY KEY,
    mode                VARCHAR(20)     NOT NULL,
    min_distance_km     DECIMAL(10,3)   NOT NULL CHECK (min_distance_km >= 0),
    max_distance_km     DECIMAL(10,3),
    rate_per_km_per_kg  DECIMAL(8,4)    NOT NULL CHECK (rate_per_km_per_kg > 0),
    effective_from      DATE            NOT NULL DEFAULT CURRENT_DATE,
    effective_to        DATE,
    is_active           BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Delivery Speed Config
CREATE TABLE IF NOT EXISTS delivery_speed_configs (
    id                      BIGSERIAL PRIMARY KEY,
    speed                   VARCHAR(20)     NOT NULL UNIQUE,
    base_courier_charge     DECIMAL(8,2)    NOT NULL DEFAULT 10.00,
    extra_charge_per_kg     DECIMAL(8,4)    NOT NULL DEFAULT 0.00,
    is_active               BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Orders
CREATE TABLE IF NOT EXISTS orders (
    id                      BIGSERIAL PRIMARY KEY,
    customer_id             BIGINT          NOT NULL REFERENCES customers(id),
    seller_id               BIGINT          NOT NULL REFERENCES sellers(id),
    nearest_warehouse_id    BIGINT          NOT NULL REFERENCES warehouses(id),
    product_id              BIGINT          NOT NULL REFERENCES products(id),
    quantity                INT             NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price              DECIMAL(10,2)   NOT NULL,
    total_product_amount    DECIMAL(12,2)   NOT NULL,
    shipping_charge         DECIMAL(10,2)   NOT NULL,
    total_amount            DECIMAL(12,2)   NOT NULL,
    delivery_speed          VARCHAR(20)     NOT NULL,
    status                  VARCHAR(30)     NOT NULL DEFAULT 'placed',
    payment_mode            VARCHAR(20)     NOT NULL DEFAULT 'prepaid',
    tracking_id             VARCHAR(100)    UNIQUE,
    estimated_delivery_date DATE,
    actual_delivery_date    DATE,
    notes                   TEXT,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Shipments
CREATE TABLE IF NOT EXISTS shipments (
    id                          BIGSERIAL PRIMARY KEY,
    order_id                    BIGINT          NOT NULL REFERENCES orders(id),
    source_warehouse_id         BIGINT          NOT NULL REFERENCES warehouses(id),
    destination_customer_id     BIGINT          NOT NULL REFERENCES customers(id),
    distance_km                 DECIMAL(10,3)   NOT NULL,
    transport_mode              VARCHAR(20)     NOT NULL,
    billable_weight_kg          DECIMAL(8,3)    NOT NULL,
    rate_per_km_per_kg          DECIMAL(8,4)    NOT NULL,
    base_courier_charge         DECIMAL(8,2)    NOT NULL,
    distance_charge             DECIMAL(10,2)   NOT NULL,
    express_charge              DECIMAL(10,2)   NOT NULL DEFAULT 0,
    total_shipping_charge       DECIMAL(10,2)   NOT NULL,
    status                      VARCHAR(20)     NOT NULL DEFAULT 'pending',
    dispatched_at               TIMESTAMP,
    delivered_at                TIMESTAMP,
    created_at                  TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_products_seller_id ON products(seller_id);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_seller_id ON orders(seller_id);
CREATE INDEX IF NOT EXISTS idx_shipments_order_id ON shipments(order_id);
CREATE INDEX IF NOT EXISTS idx_warehouses_active ON warehouses(is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_transport_rates_active ON transport_rates(is_active) WHERE is_active = TRUE;
