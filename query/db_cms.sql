/*
 Navicat Premium Data Transfer

 Source Server         : mylocal
 Source Server Type    : MySQL
 Source Server Version : 80044 (8.0.44)
 Source Host           : localhost:3306
 Source Schema         : db_cms

 Target Server Type    : MySQL
 Target Server Version : 80044 (8.0.44)
 File Encoding         : 65001

 Date: 01/12/2025 14:12:02
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for audit_logs
-- ----------------------------
DROP TABLE IF EXISTS `audit_logs`;
CREATE TABLE `audit_logs`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint UNSIGNED NULL DEFAULT NULL,
  `event_type` enum('CREATE','UPDATE','DELETE','RESTORE','LOGIN','LOGOUT') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `table_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `record_id` bigint UNSIGNED NOT NULL,
  `old_values` json NULL,
  `new_values` json NULL,
  `ip_address` varbinary(16) NULL DEFAULT NULL,
  `user_agent` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `user_id`(`user_id` ASC) USING BTREE,
  INDEX `created_at`(`created_at` ASC) USING BTREE,
  INDEX `table_name`(`table_name` ASC, `record_id` ASC) USING BTREE,
  INDEX `event_type`(`event_type` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for menu
-- ----------------------------
DROP TABLE IF EXISTS `menu`;
CREATE TABLE `menu`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `label` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `icon` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `parent_id` int UNSIGNED NULL DEFAULT NULL,
  `sort_order` smallint UNSIGNED NULL DEFAULT 0,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `deleted_by` bigint UNSIGNED NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `parent_id`(`parent_id` ASC) USING BTREE,
  INDEX `deleted_at`(`deleted_at` ASC) USING BTREE,
  CONSTRAINT `menu_ibfk_1` FOREIGN KEY (`parent_id`) REFERENCES `menu` (`id`) ON DELETE SET NULL ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for role_inheritances
-- ----------------------------
DROP TABLE IF EXISTS `role_inheritances`;
CREATE TABLE `role_inheritances`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `role_id` int UNSIGNED NOT NULL,
  `parent_role_id` int UNSIGNED NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_inherit_role`(`role_id` ASC) USING BTREE,
  INDEX `idx_inherit_parent`(`parent_role_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for role_menu
-- ----------------------------
DROP TABLE IF EXISTS `role_menu`;
CREATE TABLE `role_menu`  (
  `role_id` int UNSIGNED NOT NULL,
  `menu_id` int UNSIGNED NOT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `deleted_by` bigint UNSIGNED NULL DEFAULT NULL,
  PRIMARY KEY (`role_id`, `menu_id`) USING BTREE,
  INDEX `menu_id`(`menu_id` ASC) USING BTREE,
  INDEX `deleted_at`(`deleted_at` ASC) USING BTREE,
  CONSTRAINT `role_menu_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `role_menu_ibfk_2` FOREIGN KEY (`menu_id`) REFERENCES `menu` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for roles
-- ----------------------------
DROP TABLE IF EXISTS `roles`;
CREATE TABLE `roles`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `deleted_by` bigint UNSIGNED NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `name`(`name` ASC) USING BTREE,
  INDEX `deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for user_menu
-- ----------------------------
DROP TABLE IF EXISTS `user_menu`;
CREATE TABLE `user_menu`  (
  `user_id` bigint UNSIGNED NOT NULL,
  `menu_id` int UNSIGNED NOT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `deleted_by` bigint UNSIGNED NULL DEFAULT NULL,
  PRIMARY KEY (`user_id`, `menu_id`) USING BTREE,
  INDEX `deleted_at`(`deleted_at` ASC) USING BTREE,
  INDEX `menu_id`(`menu_id` ASC) USING BTREE,
  CONSTRAINT `user_menu_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `user_menu_ibfk_2` FOREIGN KEY (`menu_id`) REFERENCES `menu` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for user_roles
-- ----------------------------
DROP TABLE IF EXISTS `user_roles`;
CREATE TABLE `user_roles`  (
  `user_id` bigint UNSIGNED NOT NULL,
  `role_id` int UNSIGNED NOT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `deleted_by` bigint UNSIGNED NULL DEFAULT NULL,
  PRIMARY KEY (`user_id`, `role_id`) USING BTREE,
  INDEX `role_id`(`role_id` ASC) USING BTREE,
  INDEX `deleted_at`(`deleted_at` ASC) USING BTREE,
  CONSTRAINT `user_roles_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT,
  CONSTRAINT `user_roles_ibfk_2` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `email` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `password_hash` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `status` tinyint UNSIGNED NULL DEFAULT 1,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `deleted_by` bigint UNSIGNED NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `username`(`username` ASC) USING BTREE,
  UNIQUE INDEX `email`(`email` ASC) USING BTREE,
  INDEX `email_2`(`email` ASC) USING BTREE,
  INDEX `username_2`(`username` ASC) USING BTREE,
  INDEX `deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- View structure for menu_navigation
-- ----------------------------
DROP VIEW IF EXISTS `menu_navigation`;
CREATE ALGORITHM = UNDEFINED SQL SECURITY DEFINER VIEW `menu_navigation` AS with recursive `menu_tree` as (select `m`.`id` AS `parent_id`,`c`.`id` AS `child_id` from (`menu` `m` left join `menu` `c` on((`c`.`parent_id` = `m`.`id`))) where ((`m`.`deleted_at` is null) and (`c`.`deleted_at` is null)) union all select `mt`.`parent_id` AS `parent_id`,`c`.`id` AS `child_id` from (`menu` `c` join `menu_tree` `mt` on((`c`.`parent_id` = `mt`.`child_id`))) where (`c`.`deleted_at` is null)) select `m`.`id` AS `id`,`m`.`label` AS `label`,(case when ((`m`.`url` is not null) and (`m`.`url` <> '')) then `m`.`url` else 'javascript:void(0);' end) AS `url`,`m`.`icon` AS `icon`,coalesce(json_arrayagg(json_object('label',`c`.`label`,'parent_id',`c`.`parent_id`,'url',`c`.`url`)),json_array()) AS `children` from ((`menu` `m` left join `menu_tree` `mt` on((`m`.`id` = `mt`.`parent_id`))) left join `menu` `c` on((`c`.`id` = `mt`.`child_id`))) where ((`m`.`deleted_at` is null) and (`m`.`parent_id` is null)) group by `m`.`id`,`m`.`label`,`url` order by `m`.`sort_order`,`m`.`id`;

-- ----------------------------
-- View structure for v_roles
-- ----------------------------
DROP VIEW IF EXISTS `v_roles`;
CREATE ALGORITHM = UNDEFINED SQL SECURITY DEFINER VIEW `v_roles` AS with recursive `all_children` as (select `r`.`id` AS `parent_id`,`c`.`id` AS `child_id`,1 AS `level` from ((`role_inheritances` `ri` join `roles` `r` on((`r`.`id` = `ri`.`parent_role_id`))) join `roles` `c` on((`c`.`id` = `ri`.`role_id`))) union all select `ac`.`parent_id` AS `parent_id`,`c`.`id` AS `child_id`,(`ac`.`level` + 1) AS `level` from ((`role_inheritances` `ri` join `roles` `c` on((`c`.`id` = `ri`.`role_id`))) join `all_children` `ac` on((`ri`.`parent_role_id` = `ac`.`child_id`)))) select distinct `p`.`id` AS `role_id`,`p`.`name` AS `role_name`,`ac`.`child_id` AS `child_id`,`c`.`name` AS `child_name`,`ac`.`level` AS `level` from ((`all_children` `ac` join `roles` `p` on((`p`.`id` = `ac`.`parent_id`))) join `roles` `c` on((`c`.`id` = `ac`.`child_id`))) order by `role_id`,`ac`.`level`,`ac`.`child_id`;

SET FOREIGN_KEY_CHECKS = 1;
