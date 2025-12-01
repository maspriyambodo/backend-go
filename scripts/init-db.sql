-- AdminBE Database Initialization Script
-- This script creates the basic database structure for the admin backend application

USE adminbe_dev;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT UNSIGNED NULL,
    PRIMARY KEY (id),
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_deleted_at (deleted_at)
);

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NULL,
    event_type VARCHAR(50) NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    record_id BIGINT UNSIGNED NOT NULL,
    old_values JSON NULL,
    new_values JSON NULL,
    ip_address VARBINARY(16) NULL,
    user_agent VARCHAR(500) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_user_id (user_id),
    INDEX idx_event_type (event_type),
    INDEX idx_table_name (table_name),
    INDEX idx_record_id (record_id),
    INDEX idx_created_at (created_at)
);

-- Create menu table
CREATE TABLE IF NOT EXISTS menu (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    label VARCHAR(255) NOT NULL,
    url VARCHAR(500) NULL,
    icon VARCHAR(100) NULL,
    parent_id INT UNSIGNED NULL,
    sort_order SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT UNSIGNED NULL,
    PRIMARY KEY (id),
    INDEX idx_parent_id (parent_id),
    INDEX idx_sort_order (sort_order),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (parent_id) REFERENCES menu(id) ON DELETE SET NULL
);

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT UNSIGNED NULL,
    PRIMARY KEY (id),
    INDEX idx_name (name),
    INDEX idx_deleted_at (deleted_at)
);

-- Create role_inheritances table
CREATE TABLE IF NOT EXISTS role_inheritances (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    role_id INT UNSIGNED NOT NULL,
    parent_role_id INT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_role_parent (role_id, parent_role_id),
    INDEX idx_role_id (role_id),
    INDEX idx_parent_role_id (parent_role_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- Create role_menu table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS role_menu (
    role_id INT UNSIGNED NOT NULL,
    menu_id INT UNSIGNED NOT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT UNSIGNED NULL,
    PRIMARY KEY (role_id, menu_id),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (menu_id) REFERENCES menu(id) ON DELETE CASCADE
);

-- Create user_menu table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS user_menu (
    user_id BIGINT UNSIGNED NOT NULL,
    menu_id INT UNSIGNED NOT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT UNSIGNED NULL,
    PRIMARY KEY (user_id, menu_id),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (menu_id) REFERENCES menu(id) ON DELETE CASCADE
);

-- Create user_roles table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id INT UNSIGNED NOT NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by BIGINT UNSIGNED NULL,
    PRIMARY KEY (user_id, role_id),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- Create views for hierarchical data

-- View for role hierarchy (v_roles)
CREATE OR REPLACE VIEW v_roles AS
WITH RECURSIVE role_hierarchy AS (
    -- Base case: roles with no parents
    SELECT
        id as role_id,
        name as role_name,
        id as child_id,
        name as child_name,
        0 as level
    FROM roles
    WHERE deleted_at IS NULL

    UNION ALL

    -- Recursive case: find children
    SELECT
        rh.role_id,
        rh.role_name,
        r.id as child_id,
        r.name as child_name,
        rh.level + 1 as level
    FROM role_hierarchy rh
    JOIN role_inheritances ri ON rh.child_id = ri.parent_role_id
    JOIN roles r ON ri.role_id = r.id
    WHERE r.deleted_at IS NULL
)
SELECT * FROM role_hierarchy ORDER BY role_id, level, child_id;

-- View for menu navigation (menu_navigation)
CREATE OR REPLACE VIEW menu_navigation AS
WITH RECURSIVE menu_tree AS (
    -- Base case: root menus (no parent)
    SELECT
        id,
        label,
        COALESCE(url, '') as url,
        COALESCE(icon, '') as icon,
        CAST('[]' AS JSON) as children,
        sort_order
    FROM menu
    WHERE parent_id IS NULL AND deleted_at IS NULL

    UNION ALL

    -- Recursive case: build children
    SELECT
        m.id,
        m.label,
        COALESCE(m.url, '') as url,
        COALESCE(m.icon, '') as icon,
        JSON_ARRAYAGG(
            JSON_OBJECT(
                'id', mt.id,
                'label', mt.label,
                'url', mt.url,
                'icon', mt.icon,
                'children', mt.children
            )
        ) as children,
        m.sort_order
    FROM menu m
    JOIN menu_tree mt ON m.id = mt.id
    WHERE m.parent_id IS NOT NULL AND m.deleted_at IS NULL
    GROUP BY m.id, m.label, m.url, m.icon, m.sort_order
)
SELECT
    id,
    label,
    url,
    icon,
    children
FROM menu_tree
ORDER BY sort_order;

-- Insert default admin user (password: admin123)
-- Password hash generated for 'admin123' - in production, use proper password hashing
INSERT IGNORE INTO users (username, email, password_hash, status) VALUES
('admin', 'admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 1);

-- Insert default roles
INSERT IGNORE INTO roles (name, description) VALUES
('admin', 'Administrator with full access'),
('user', 'Regular user'),
('moderator', 'Content moderator');

-- Insert default menus
INSERT IGNORE INTO menu (label, url, icon, parent_id, sort_order) VALUES
('Dashboard', '/dashboard', 'dashboard', NULL, 1),
('Users', '/users', 'users', NULL, 2),
('Roles', '/roles', 'roles', NULL, 3),
('Menu', '/menu', 'menu', NULL, 4),
('Audit Logs', '/audit-logs', 'logs', NULL, 5);

-- Insert default role-menu permissions (admin gets all menus)
INSERT IGNORE INTO role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM roles r, menu m
WHERE r.name = 'admin' AND m.deleted_at IS NULL;

-- Assign admin role to admin user
INSERT IGNORE INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE u.username = 'admin' AND r.name = 'admin';

COMMIT;
