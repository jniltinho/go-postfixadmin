-- Migrate existing admins: assign domain_admin role to every active non-superadmin
-- that has entries in domain_admins but no row in rbac_admin_roles yet.
--
-- Safe to run multiple times (INSERT IGNORE).
-- Run AFTER: postfixadmin migrate rbac
--
-- Usage (MariaDB/MySQL):
--   mariadb -u postfixadmin -p postfixadmin < rbac_migrate_existing_admins.sql

INSERT IGNORE INTO rbac_admin_roles (username, role_id, domain, created)
SELECT
    da.username,
    (SELECT id FROM rbac_roles WHERE name = 'domain_admin'),
    '',
    NOW()
FROM (
    SELECT DISTINCT da.username
    FROM domain_admins da
    JOIN admin a ON a.username = da.username
    WHERE a.active      = 1
      AND a.superadmin  = 0
      AND da.active     = 1
) AS da
WHERE NOT EXISTS (
    SELECT 1 FROM rbac_admin_roles ar WHERE ar.username = da.username
);

-- Show what was assigned
SELECT
    ar.username,
    r.name  AS role,
    ar.domain,
    ar.created
FROM rbac_admin_roles ar
JOIN rbac_roles r ON r.id = ar.role_id
ORDER BY ar.created DESC;
