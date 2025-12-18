-- =====================================================
-- Migration 000036: Disable pgvector Extension (Down)
-- =====================================================

DROP EXTENSION IF EXISTS vector;
