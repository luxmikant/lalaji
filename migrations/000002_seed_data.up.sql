-- Seed: Warehouses
INSERT INTO warehouses (name, code, lat, lng, city, state, pincode, contact_person, contact_phone, max_capacity_sqft, is_active)
VALUES
    ('BLR_Warehouse',   'BLR',   12.9716,   77.5946,  'Bengaluru',  'Karnataka',    '560001', 'Raj Kumar',    '9876543210', 50000, TRUE),
    ('MUMB_Warehouse',  'MUMB',  19.0760,   72.8777,  'Mumbai',     'Maharashtra',  '400001', 'Priya Shah',   '9876543211', 75000, TRUE),
    ('DEL_Warehouse',   'DEL',   28.7041,   77.1025,  'New Delhi',  'Delhi',        '110001', 'Amit Gupta',   '9876543212', 60000, TRUE),
    ('HYD_Warehouse',   'HYD',   17.3850,   78.4867,  'Hyderabad',  'Telangana',    '500001', 'Sita Reddy',   '9876543213', 45000, TRUE),
    ('CHN_Warehouse',   'CHN',   13.0827,   80.2707,  'Chennai',    'Tamil Nadu',   '600001', 'Karthik R',    '9876543214', 40000, TRUE)
ON CONFLICT (name) DO NOTHING;

-- Seed: Transport Rates
INSERT INTO transport_rates (mode, min_distance_km, max_distance_km, rate_per_km_per_kg, is_active)
VALUES
    ('minivan',    0,    100,  3.0000, TRUE),
    ('truck',      100,  500,  2.0000, TRUE),
    ('aeroplane',  500,  NULL, 1.0000, TRUE)
ON CONFLICT DO NOTHING;

-- Seed: Delivery Speed Configs
INSERT INTO delivery_speed_configs (speed, base_courier_charge, extra_charge_per_kg, is_active)
VALUES
    ('standard', 10.00, 0.0000, TRUE),
    ('express',  10.00, 1.2000, TRUE)
ON CONFLICT (speed) DO NOTHING;

-- Seed: Sellers
INSERT INTO sellers (name, owner_name, phone, email, gst_number, business_type, lat, lng, city, state, pincode, rating, is_verified, is_active)
VALUES
    ('Nestle India Distributor', 'Rahul Mehta',   '9100000001', 'nestle@example.com',  '29AABCU9603R1ZM', 'distributor',  12.9352, 77.6245, 'Bengaluru', 'Karnataka',   '560037', 4.50, TRUE, TRUE),
    ('Premium Rice Traders',     'Suresh Patil',  '9100000002', 'rice@example.com',    '27AADCS2365R1Z1', 'wholesaler',   19.1136, 72.8697, 'Mumbai',    'Maharashtra', '400053', 4.20, TRUE, TRUE),
    ('Gujarat Sugar Mills',      'Bhavesh Patel', '9100000003', 'sugar@example.com',   '24AAACG1234R1ZP', 'manufacturer', 23.0225, 72.5714, 'Ahmedabad', 'Gujarat',     '380001', 4.70, TRUE, TRUE)
ON CONFLICT (phone) DO NOTHING;

-- Seed: Products
INSERT INTO products (seller_id, name, description, sku, category, mrp, selling_price, actual_weight_kg, length_cm, width_cm, height_cm, is_fragile, is_perishable, stock_quantity, min_order_quantity, is_active)
VALUES
    (1, 'Maggi 500g Packet',     '500g Maggi noodles Family Pack',  'NES-MAG-500',  'snacks',  14.00,  10.00,  0.500,  10.00,  10.00,  10.00,  FALSE, FALSE, 10000, 24,  TRUE),
    (1, 'Nescafe Classic 200g',  '200g instant coffee jar',         'NES-COF-200',  'beverages', 350.00, 280.00, 0.250,  8.00,   8.00,   12.00,  TRUE,  FALSE, 5000,  12,  TRUE),
    (2, 'Basmati Rice 10Kg',     'Premium aged basmati rice bag',   'PRT-RIC-10K',  'rice',    700.00, 500.00, 10.000, 60.00,  40.00,  15.00,  FALSE, FALSE, 2000,  5,   TRUE),
    (2, 'Sona Masoori Rice 25Kg','25Kg Sona Masoori rice',          'PRT-RIC-25K',  'rice',    1200.00, 950.00, 25.000, 80.00,  50.00,  20.00, FALSE, FALSE, 1000,  3,   TRUE),
    (3, 'White Sugar 25Kg',      '25Kg refined white sugar bag',    'GSM-SUG-25K',  'sugar',   900.00, 700.00, 25.000, 70.00,  45.00,  25.00,  FALSE, FALSE, 3000,  5,   TRUE),
    (3, 'Jaggery Powder 5Kg',    '5Kg organic jaggery powder',      'GSM-JAG-5K',   'sugar',   400.00, 320.00, 5.000,  30.00,  20.00,  15.00,  FALSE, FALSE, 1500,  10,  TRUE)
ON CONFLICT (seller_id, sku) DO NOTHING;

-- Seed: Customers (Kirana Stores)
INSERT INTO customers (name, owner_name, phone, email, gst_number, store_type, credit_limit, preferred_delivery_slot, lat, lng, address_line1, city, state, pincode, is_active)
VALUES
    ('Shree Kirana Store',        'Ramesh Kumar',    '9847000001', 'shree@example.com',   '29AABCS1234R1Z1', 'grocery',  50000.00, 'morning',  12.9716,  77.5946,  'MG Road, Koramangala',       'Bengaluru',  'Karnataka',    '560034', TRUE),
    ('Andheri Mini Mart',         'Sunil Sharma',    '9847000002', 'andheri@example.com',  '27AABCA5678R1Z2', 'general',  75000.00, 'evening',  19.1197,  72.8464,  'Andheri West, Link Road',    'Mumbai',     'Maharashtra',  '400053', TRUE),
    ('Dilli Grocery Hub',         'Pankaj Verma',    '9847000003', 'dilli@example.com',    '07AABCD9012R1Z3', 'grocery',  100000.00, 'morning', 28.6139,  77.2090,  'Connaught Place',            'New Delhi',  'Delhi',        '110001', TRUE),
    ('Hyderabad Fresh Mart',      'Lakshmi Reddy',   '9847000004', 'hyd@example.com',     '36AABCE3456R1Z4', 'dairy',    40000.00,  'afternoon', 17.4065, 78.4772, 'Banjara Hills, Road No 1',   'Hyderabad',  'Telangana',    '500034', TRUE),
    ('Chennai Bazaar',            'Murugan S',       '9847000005', 'chennai@example.com',  '33AABCF7890R1Z5', 'general',  60000.00,  'morning',  13.0569, 80.2425, 'T Nagar, Usman Road',        'Chennai',    'Tamil Nadu',   '600017', TRUE)
ON CONFLICT (phone) DO NOTHING;
