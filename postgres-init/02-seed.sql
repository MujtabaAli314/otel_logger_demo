-- ============================================================================
-- Seed data for the simplified FIN schema (users, transactions only)
-- ============================================================================

INSERT INTO users (name, email) VALUES
    ('Alice Anderson', 'alice@example.com'),
    ('Bob Brown',       'bob@example.com'),
    ('Carol Clarke',    'carol@example.com');

INSERT INTO transactions (user_id, amount, currency, type, merchant, description, created_at)
SELECT u.id, v.amount, v.currency, v.type, v.merchant, v.description, v.created_at
FROM (VALUES
    ('alice@example.com', 1200.00, 'USD', 'credit', 'Acme Payroll',      'Monthly salary',              '2026-07-01 09:00:00+00'::timestamptz),
    ('alice@example.com',  -89.99, 'USD', 'debit',  'Streamly',          'Monthly subscription',        '2026-07-03 14:22:00+00'),
    ('alice@example.com', -300.00, 'USD', 'debit',  'ATM Withdrawal',    'Cash withdrawal',              '2026-07-05 08:15:00+00'),
    ('alice@example.com',  -45.50, 'USD', 'debit',  'FreshMart',         'Grocery purchase',             '2026-07-08 17:30:00+00'),
    ('bob@example.com',   -1200.00, 'USD', 'debit',  'Internal Transfer', 'Transfer to savings',         '2026-07-07 11:45:00+00'),
    ('bob@example.com',    800.00, 'USD', 'credit', 'ClientCo',          'Freelance payment received',  '2026-07-14 16:40:00+00'),
    ('bob@example.com',    -60.00, 'USD', 'debit',  'GasStop',           'Card payment at pump',        '2026-07-15 07:20:00+00'),
    ('bob@example.com',   -500.00, 'USD', 'debit',  'RentCo',            'Monthly rent payment',        '2026-07-01 06:00:00+00'),
    ('carol@example.com',  150.00, 'EUR', 'credit', 'TechGadgets',       'Refund for returned item',    '2026-07-10 13:05:00+00'),
    ('carol@example.com',  -22.10, 'EUR', 'debit',  'Bean There',        'Coffee shop purchase',        '2026-07-13 08:05:00+00'),
    ('carol@example.com', 2000.00, 'EUR', 'credit', 'Acme Payroll',      'Monthly salary',              '2026-07-01 09:00:00+00'),
    ('carol@example.com',  -75.25, 'EUR', 'debit',  'CityMart',          'Weekly grocery run',          '2026-07-20 18:12:00+00')
) AS v(email, amount, currency, type, merchant, description, created_at)
JOIN users u ON u.email = v.email;
