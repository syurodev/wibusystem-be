-- =====================================================
-- Migration 000036: Enable pgvector Extension
-- Description: Install pgvector for vector similarity search
-- =====================================================

CREATE EXTENSION IF NOT EXISTS vector;
