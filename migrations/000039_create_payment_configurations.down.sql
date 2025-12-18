-- Migration: Drop Payment Configuration Table
-- Description: Rollback for payment configurations

DROP TABLE IF EXISTS payment.configurations;
