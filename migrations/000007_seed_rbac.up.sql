INSERT INTO roles (id, name, description)
VALUES
    (gen_random_uuid(), 'admin', 'Administrator with full access')
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles (id, name, description)
VALUES
    (gen_random_uuid(), 'user', 'Standard user')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (id, name, description) VALUES
    (gen_random_uuid(), 'users.view', 'View users'),
    (gen_random_uuid(), 'users.create', 'Create users'),
    (gen_random_uuid(), 'users.update', 'Update users'),
    (gen_random_uuid(), 'users.delete', 'Delete users'),
    (gen_random_uuid(), 'roles.view', 'View roles'),
    (gen_random_uuid(), 'roles.manage', 'Manage roles'),
    (gen_random_uuid(), 'permissions.view', 'View permissions'),
    (gen_random_uuid(), 'permissions.manage', 'Manage permissions')
ON CONFLICT (name) DO NOTHING;

-- Admin gets every permission.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- User gets view-only permissions.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('users.view', 'roles.view', 'permissions.view')
WHERE r.name = 'user'
ON CONFLICT DO NOTHING;
