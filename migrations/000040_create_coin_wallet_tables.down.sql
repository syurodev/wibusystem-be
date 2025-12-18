-- Migration: Drop Coin Wallet Tables
-- Description: Rollback for coin wallet tables

DROP TABLE IF EXISTS payment.transactions;
DROP TABLE IF EXISTS payment.topup_orders;
DROP TABLE IF EXISTS payment.coin_packages;
DROP TABLE IF EXISTS payment.user_wallets;

DROP TYPE IF EXISTS payment.transaction_type;
DROP TYPE IF EXISTS payment.topup_status;
