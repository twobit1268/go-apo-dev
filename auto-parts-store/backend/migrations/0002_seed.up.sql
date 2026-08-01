INSERT INTO categories (id, name, slug) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Brakes', 'brakes'),
    ('22222222-2222-2222-2222-222222222222', 'Filters', 'filters'),
    ('33333333-3333-3333-3333-333333333333', 'Batteries', 'batteries');

INSERT INTO parts (id, sku, name, description, category_id, price_cents, stock_qty) VALUES
    ('aaaaaaaa-0001-0001-0001-000000000001', 'BRK-1001', 'Ceramic Brake Pad Set (Front)',
        'Low-dust ceramic pads for most sedans and crossovers.',
        '11111111-1111-1111-1111-111111111111', 4899, 50),
    ('aaaaaaaa-0001-0001-0001-000000000002', 'BRK-1002', 'Drilled & Slotted Rotor (Pair)',
        'Improved heat dissipation for spirited driving.',
        '11111111-1111-1111-1111-111111111111', 12999, 20),
    ('aaaaaaaa-0002-0002-0002-000000000001', 'FIL-2001', 'Engine Oil Filter',
        'Spin-on oil filter, fits most 4-cylinder engines.',
        '22222222-2222-2222-2222-222222222222', 899, 200),
    ('aaaaaaaa-0002-0002-0002-000000000002', 'FIL-2002', 'Cabin Air Filter',
        'Activated-carbon cabin filter for cleaner interior air.',
        '22222222-2222-2222-2222-222222222222', 1499, 150),
    ('aaaaaaaa-0003-0003-0003-000000000001', 'BAT-3001', '12V Automotive Battery (Group 35)',
        '650 CCA, 3-year warranty, maintenance-free.',
        '33333333-3333-3333-3333-333333333333', 15999, 30);
