INSERT INTO users (id, email, role, created_at, password_hash)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin@dummy.local', 'admin', NOW() AT TIME ZONE 'utc', NULL),
    ('00000000-0000-0000-0000-000000000002', 'user@dummy.local', 'user', NOW() AT TIME ZONE 'utc', NULL)
    ON CONFLICT (id) DO NOTHING;
