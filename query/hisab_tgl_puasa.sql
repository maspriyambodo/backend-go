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

 Date: 10/12/2025 22:54:57
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for hisab_tgl_puasa
-- ----------------------------
DROP TABLE IF EXISTS `hisab_tgl_puasa`;
CREATE TABLE `hisab_tgl_puasa`  (
  `tgl_id` int NOT NULL AUTO_INCREMENT,
  `tgl_tahun` int NULL DEFAULT NULL,
  `tgl_start` date NULL DEFAULT NULL,
  `tgl_end` date NULL DEFAULT NULL,
  `tgl_status` int NULL DEFAULT 0,
  `tgl_hijriah` int NULL DEFAULT NULL,
  `time_add` datetime NULL DEFAULT NULL,
  `time_update` datetime NULL DEFAULT NULL,
  `user_add` int NULL DEFAULT NULL,
  `user_update` int NULL DEFAULT NULL,
  PRIMARY KEY (`tgl_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 14 CHARACTER SET = latin1 COLLATE = latin1_swedish_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
