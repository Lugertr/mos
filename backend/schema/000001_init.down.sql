-- --------------------
-- 1) DROP TRIGGERS (если есть)
-- --------------------
DROP TRIGGER IF EXISTS trg_documents_validate_and_fill_geom ON documents;
DROP TRIGGER IF EXISTS trg_log_users ON users;
DROP TRIGGER IF EXISTS trg_log_documents ON documents;
DROP TRIGGER IF EXISTS trg_log_document_tags ON document_tags;
DROP TRIGGER IF EXISTS trg_log_tags ON tags;
DROP TRIGGER IF EXISTS trg_log_authors ON authors;
DROP TRIGGER IF EXISTS trg_log_document_types ON document_types;
DROP TRIGGER IF EXISTS trg_after_delete_document_cleanup ON documents;

-- also any generic trg_log_* created in compact variant:
DO $$
DECLARE t record;
BEGIN
  FOR t IN
    SELECT tgname, relname
    FROM pg_trigger tr JOIN pg_class c ON tr.tgrelid = c.oid
    WHERE tgname LIKE 'trg_log_%'
  LOOP
    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I', t.tgname, t.relname);
  END LOOP;
END $$;

-- --------------------
-- 2) DROP FUNCTIONS (триgger functions и вспомогательные)
-- --------------------
-- триггерные и валидация
DROP FUNCTION IF EXISTS trg_documents_validate_and_fill_geom() CASCADE;
DROP FUNCTION IF EXISTS fn_validate_geojson(JSONB) CASCADE;
DROP FUNCTION IF EXISTS trg_after_delete_document_cleanup() CASCADE;
DROP FUNCTION IF EXISTS fn_log_changes() CASCADE;

-- внутренние helper'ы
DROP FUNCTION IF EXISTS _internal_get_or_create_author(TEXT) CASCADE;
DROP FUNCTION IF EXISTS _internal_get_or_create_tag(TEXT) CASCADE;
DROP FUNCTION IF EXISTS _internal_attach_tag_to_document(INTEGER, TEXT) CASCADE;
DROP FUNCTION IF EXISTS _internal_cleanup_unused_tags() CASCADE;
DROP FUNCTION IF EXISTS _internal_cleanup_unused_authors() CASCADE;

-- auth / users / admin
DROP FUNCTION IF EXISTS fn_register_user(TEXT, TEXT, TEXT, SMALLINT) CASCADE;
DROP FUNCTION IF EXISTS fn_authorize_user(TEXT, TEXT) CASCADE;
DROP FUNCTION IF EXISTS is_user_admin(INTEGER) CASCADE;

-- permission checks
DROP FUNCTION IF EXISTS _can_user_edit_document(INTEGER, INTEGER) CASCADE;
DROP FUNCTION IF EXISTS _can_user_view_document(INTEGER, INTEGER) CASCADE;

-- CRUD documents
DROP FUNCTION IF EXISTS fn_add_document(INTEGER, TEXT, DATE, INTEGER, TEXT, INTEGER, BYTEA, JSONB, TEXT[], TEXT) CASCADE;
DROP FUNCTION IF EXISTS fn_update_document(INTEGER, INTEGER, TEXT, DATE, INTEGER, TEXT, INTEGER, BYTEA, JSONB, TEXT[], TEXT) CASCADE;
DROP FUNCTION IF EXISTS fn_delete_document(INTEGER, INTEGER) CASCADE;

-- permissions management
DROP FUNCTION IF EXISTS fn_set_document_permission(INTEGER, INTEGER, INTEGER, BOOLEAN, BOOLEAN) CASCADE;
DROP FUNCTION IF EXISTS fn_remove_document_permission(INTEGER, INTEGER, INTEGER) CASCADE;

-- searches / getters
DROP FUNCTION IF EXISTS fn_get_documents_by_tag_secure(TEXT, TEXT, TEXT, DATE, DATE, INTEGER) CASCADE;
DROP FUNCTION IF EXISTS fn_get_document_secure(INTEGER, INTEGER) CASCADE;

-- logs readers
DROP FUNCTION IF EXISTS fn_get_logs_by_user(INTEGER, INTEGER, TIMESTAMPTZ, TIMESTAMPTZ) CASCADE;
DROP FUNCTION IF EXISTS fn_get_logs_by_table(INTEGER, TEXT, TIMESTAMPTZ, TIMESTAMPTZ) CASCADE;
DROP FUNCTION IF EXISTS fn_get_logs_by_date(INTEGER, TIMESTAMPTZ, TIMESTAMPTZ) CASCADE;

-- cleanup / scheduler
DROP FUNCTION IF EXISTS fn_periodic_cleanup() CASCADE;

-- --------------------
-- 3) DROP INDEXES (если хотите удалить отдельно; можно пропустить — DROP TABLE ... CASCADE удалит их)
-- --------------------
DROP INDEX IF EXISTS documents_title_idx;
DROP INDEX IF EXISTS documents_title_trgm_idx;
DROP INDEX IF EXISTS documents_created_at_idx;
DROP INDEX IF EXISTS documents_document_date_idx;
DROP INDEX IF EXISTS documents_geom_gist;
DROP INDEX IF EXISTS tags_name_lower_idx;
DROP INDEX IF EXISTS authors_name_lower_idx;
DROP INDEX IF EXISTS doc_types_name_idx;
DROP INDEX IF EXISTS document_tags_tag_idx;
DROP INDEX IF EXISTS document_tags_document_idx;
DROP INDEX IF EXISTS documents_type_author_created_idx;
DROP INDEX IF EXISTS document_permissions_user_idx;
DROP INDEX IF EXISTS logs_action_time_idx;
DROP INDEX IF EXISTS logs_user_id_idx;

-- --------------------
-- 4) DROP TABLES (в безопасном порядке)
-- --------------------
DROP TABLE IF EXISTS document_permissions CASCADE;
DROP TABLE IF EXISTS document_tags CASCADE;
DROP TABLE IF EXISTS logs CASCADE;
DROP TABLE IF EXISTS documents CASCADE;
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS authors CASCADE;
DROP TABLE IF EXISTS document_types CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS roles CASCADE;

-- --------------------
-- 5) DROP TYPES
-- --------------------
DROP TYPE IF EXISTS privacy_type CASCADE;
DROP TYPE IF EXISTS action_type CASCADE;

-- --------------------
-- 6) DROP SEQUENCES (обычно удаляются через DROP TABLE ... CASCADE, но на всякий случай)
-- --------------------
DO $$
DECLARE s record;
BEGIN
  FOR s IN SELECT sequence_schema, sequence_name FROM information_schema.sequences WHERE sequence_schema = 'public' LOOP
    BEGIN
      EXECUTE format('DROP SEQUENCE IF EXISTS %I.%I CASCADE', s.sequence_schema, s.sequence_name);
    EXCEPTION WHEN OTHERS THEN
      -- ignore sequences we cannot drop
      RAISE NOTICE 'Skipping sequence %I.%I: %', s.sequence_schema, s.sequence_name, SQLERRM;
    END;
  END LOOP;
END $$;

-- --------------------
-- 7) DROP ROLES (если хотите)
-- --------------------
DROP ROLE IF EXISTS app_service;
DROP ROLE IF EXISTS db_admin;

-- --------------------
-- 8) DROP EXTENSIONS (в конце)
-- --------------------
DROP EXTENSION IF EXISTS pg_trgm CASCADE;
DROP EXTENSION IF EXISTS citext CASCADE;
DROP EXTENSION IF EXISTS postgis CASCADE;
DROP EXTENSION IF EXISTS pgcrypto CASCADE;

-- --------------------
-- 9) NOTICE
-- --------------------
DO $$ BEGIN RAISE NOTICE 'All requested objects DROPped (IF EXISTS).'; END $$;
