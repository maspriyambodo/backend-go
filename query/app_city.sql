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

 Date: 10/12/2025 22:54:36
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for app_city
-- ----------------------------
DROP TABLE IF EXISTS `app_city`;
CREATE TABLE `app_city`  (
  `city_id` int NOT NULL AUTO_INCREMENT,
  `city_title` varchar(40) CHARACTER SET latin1 COLLATE latin1_swedish_ci NULL DEFAULT NULL,
  `city_province` int NOT NULL,
  `city_id_new` int NOT NULL,
  PRIMARY KEY (`city_id`) USING BTREE
) ENGINE = MyISAM AUTO_INCREMENT = 523 CHARACTER SET = latin1 COLLATE = latin1_swedish_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
