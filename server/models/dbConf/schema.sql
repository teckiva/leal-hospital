-- Lael Hospital Database Schema

-- Users table (Admin and Staff)
CREATE TABLE `lael_users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `mobile` varchar(15) NOT NULL,
  `email` varchar(255) NOT NULL,
  `designation` enum('doctor','nurse','staff') NOT NULL,
  `status` enum('active','inactive','temporary_inactive') NOT NULL DEFAULT 'active',
  `is_admin` tinyint(1) NOT NULL DEFAULT '0',
  `is_approved` tinyint(1) NOT NULL DEFAULT '0',
  `approved_by` bigint DEFAULT NULL,
  `password_hash` varchar(255) DEFAULT NULL,
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `last_login_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `mobile` (`mobile`),
  UNIQUE KEY `email` (`email`),
  KEY `idx_mobile` (`mobile`),
  KEY `idx_email` (`email`),
  KEY `idx_is_admin` (`is_admin`),
  KEY `approved_by` (`approved_by`),
  CONSTRAINT `lael_users_ibfk_1` FOREIGN KEY (`approved_by`) REFERENCES `lael_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- OTP table
CREATE TABLE `lael_otp` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `mobile` varchar(15) NOT NULL,
  `email` varchar(255) NOT NULL,
  `otp` varchar(6) NOT NULL,
  `expiry` datetime NOT NULL,
  `is_validated` tinyint(1) NOT NULL DEFAULT '0',
  `otp_type` enum('registration','login','forgot_password') NOT NULL,
  `retry_count` int NOT NULL DEFAULT '0',
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_mobile_type` (`mobile`,`otp_type`),
  KEY `idx_email_type` (`email`,`otp_type`),
  KEY `idx_expiry` (`expiry`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Patients table
CREATE TABLE `lael_patients` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `mobile` varchar(15) NOT NULL,
  `opd_id` varchar(50) NOT NULL,
  `age` int NOT NULL,
  `sex` enum('male','female','other') NOT NULL,
  `address_locality` varchar(255) DEFAULT NULL,
  `address_city` varchar(100) DEFAULT NULL,
  `address_state` varchar(100) DEFAULT NULL,
  `address_pincode` varchar(10) DEFAULT NULL,
  `visit_number` int NOT NULL DEFAULT '1',
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `opd_id` (`opd_id`),
  KEY `idx_mobile` (`mobile`),
  KEY `idx_opd_id` (`opd_id`),
  KEY `idx_created_on` (`created_on`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Patient OPD records
CREATE TABLE `patient_opd` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `patient_id` bigint NOT NULL,
  `doctor_id` bigint NOT NULL,
  `symptoms` json DEFAULT NULL,
  `prescription` json DEFAULT NULL,
  `medicines` json DEFAULT NULL,
  `future_suggestion` json DEFAULT NULL,
  `template_version` int NOT NULL DEFAULT '1',
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_patient_id` (`patient_id`),
  KEY `idx_doctor_id` (`doctor_id`),
  KEY `idx_created_on` (`created_on`),
  CONSTRAINT `patient_opd_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `lael_patients` (`id`),
  CONSTRAINT `patient_opd_ibfk_2` FOREIGN KEY (`doctor_id`) REFERENCES `lael_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Sessions table
CREATE TABLE `lael_sessions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `session_token` varchar(255) NOT NULL,
  `session_expiry` datetime NOT NULL,
  `last_active_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `device_info` varchar(500) DEFAULT NULL,
  `ip_address` varchar(45) DEFAULT NULL,
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `session_token` (`session_token`),
  KEY `idx_session_token` (`session_token`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_session_expiry` (`session_expiry`),
  CONSTRAINT `lael_sessions_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `lael_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Activity logs table
CREATE TABLE `lael_activity_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint DEFAULT NULL,
  `activity_type` enum('login','logout','patient_creation','staff_approval','opd_generation') NOT NULL,
  `description` text,
  `ip_address` varchar(45) DEFAULT NULL,
  `created_on` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_activity_type` (`activity_type`),
  KEY `idx_created_on` (`created_on`),
  CONSTRAINT `lael_activity_logs_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `lael_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
