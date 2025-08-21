-- Расширения
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type t WHERE t.typname = 'privacy_type') THEN
        EXECUTE 'CREATE TYPE privacy_type AS ENUM (''public'', ''private'')';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type t WHERE t.typname = 'action_type') THEN
        EXECUTE 'CREATE TYPE action_type AS ENUM (''create'', ''update'', ''delete'')';
    END IF;
END
$$;

-- Справочники и пользователи/роли
CREATE TABLE IF NOT EXISTS roles (id SMALLINT PRIMARY KEY, name TEXT NOT NULL UNIQUE);
INSERT INTO roles (id,name) SELECT * FROM (VALUES (1,'administrator'),(2,'user')) v(id,name) ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  role_id SMALLINT NOT NULL REFERENCES roles(id),
  login citext NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  full_name TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT users_login_not_blank CHECK (btrim(login::text) <> '')
);

CREATE TABLE IF NOT EXISTS authors (id SERIAL PRIMARY KEY, full_name citext NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS document_types (id SERIAL PRIMARY KEY, name citext NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS tags (id SERIAL PRIMARY KEY, name citext NOT NULL UNIQUE);

-- Документы и связанные сущности
CREATE TABLE IF NOT EXISTS documents (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  privacy privacy_type NOT NULL DEFAULT 'public',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by INTEGER REFERENCES users(id),
  updated_at TIMESTAMPTZ,
  updated_by INTEGER REFERENCES users(id),
  document_date DATE,
  author_id INTEGER REFERENCES authors(id),
  type_id INTEGER REFERENCES document_types(id),
  file_bytea BYTEA,
  geojson JSONB,
  geom geometry(Geometry,4326),
  CONSTRAINT documents_title_not_blank CHECK (btrim(title) <> '')
);

CREATE TABLE IF NOT EXISTS document_permissions (
  document_id INTEGER REFERENCES documents(id) ON DELETE CASCADE,
  user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
  can_view BOOLEAN DEFAULT FALSE,
  can_edit BOOLEAN DEFAULT FALSE,
  PRIMARY KEY (document_id, user_id)
);

CREATE TABLE IF NOT EXISTS document_tags (
  document_id INTEGER REFERENCES documents(id) ON DELETE CASCADE,
  tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (document_id, tag_id)
);

CREATE TABLE IF NOT EXISTS logs (
  id SERIAL PRIMARY KEY,
  action action_type NOT NULL,
  table_name TEXT NOT NULL,
  record_id INTEGER,
  user_id INTEGER,
  tg_op TEXT,
  session_of_user TEXT,
  user_login TEXT,
  action_time TIMESTAMPTZ NOT NULL DEFAULT now(),
  changes JSONB
);

-- Индексы (ключевые)
CREATE INDEX IF NOT EXISTS documents_title_idx         ON documents (lower(title));
CREATE INDEX IF NOT EXISTS documents_title_trgm_idx    ON documents USING gin (lower(title) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS documents_created_at_idx    ON documents (created_at);
CREATE INDEX IF NOT EXISTS documents_document_date_idx ON documents (document_date);
CREATE INDEX IF NOT EXISTS documents_geom_gist         ON documents USING GIST (geom);
CREATE INDEX IF NOT EXISTS tags_name_lower_idx         ON tags (lower(name));
CREATE INDEX IF NOT EXISTS authors_name_lower_idx      ON authors (lower(full_name));
CREATE INDEX IF NOT EXISTS doc_types_name_idx          ON document_types (lower(name));
CREATE INDEX IF NOT EXISTS document_tags_tag_idx       ON document_tags (tag_id);
CREATE INDEX IF NOT EXISTS document_permissions_user_idx ON document_permissions (user_id);
CREATE INDEX IF NOT EXISTS logs_action_time_idx        ON logs (action_time);
CREATE INDEX IF NOT EXISTS logs_user_id_idx            ON logs (user_id);

-- ========== GEOJSON: валидация и заполнение geom ==========
CREATE OR REPLACE FUNCTION fn_validate_geojson(p_geojson JSONB) RETURNS BOOLEAN
LANGUAGE plpgsql STABLE AS $$
DECLARE g geometry;
BEGIN
  IF p_geojson IS NULL THEN RETURN TRUE; END IF;
  CASE lower(coalesce(p_geojson->>'type','')) 
    WHEN '' THEN RETURN FALSE;
    WHEN 'point','linestring','polygon','multipoint','multilinestring','multipolygon','geometrycollection' THEN
      BEGIN g := ST_SetSRID(ST_GeomFromGeoJSON(p_geojson::text),4326); RETURN ST_IsValid(g); EXCEPTION WHEN OTHERS THEN RETURN FALSE; END;
    WHEN 'feature' THEN
      IF p_geojson->'geometry' IS NULL THEN RETURN FALSE; END IF;
      BEGIN g := ST_SetSRID(ST_GeomFromGeoJSON((p_geojson->'geometry')::text),4326); RETURN ST_IsValid(g); EXCEPTION WHEN OTHERS THEN RETURN FALSE; END;
    WHEN 'featurecollection' THEN
      IF p_geojson->'features' IS NULL OR jsonb_typeof(p_geojson->'features') <> 'array' THEN RETURN FALSE; END IF;
      BEGIN
        RETURN ST_IsValid(ST_Collect(ARRAY(
          SELECT ST_SetSRID(ST_GeomFromGeoJSON((f->'geometry')::text),4326)
          FROM jsonb_array_elements(p_geojson->'features') AS arr(f)
        )));
      EXCEPTION WHEN OTHERS THEN RETURN FALSE; END;
    ELSE RETURN FALSE;
  END CASE;
END;
$$;

CREATE OR REPLACE FUNCTION trg_documents_validate_and_fill_geom() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.geojson IS NULL THEN NEW.geom := NULL; RETURN NEW; END IF;
  IF NOT fn_validate_geojson(NEW.geojson) THEN RAISE EXCEPTION 'Invalid GeoJSON for document %', COALESCE(NEW.title,'unknown'); END IF;
  BEGIN
    NEW.geom := ST_SetSRID(ST_GeomFromGeoJSON(NEW.geojson::text),4326);
  EXCEPTION WHEN OTHERS THEN
    IF lower(coalesce(NEW.geojson->>'type',''))='feature' THEN
      BEGIN NEW.geom := ST_SetSRID(ST_GeomFromGeoJSON((NEW.geojson->'geometry')::text),4326); EXCEPTION WHEN OTHERS THEN NEW.geom := NULL; END;
    ELSIF lower(coalesce(NEW.geojson->>'type',''))='featurecollection' THEN
      BEGIN NEW.geom := (SELECT ST_Collect(array_agg(g)) FROM (SELECT ST_SetSRID(ST_GeomFromGeoJSON((f->'geometry')::text),4326) AS g FROM jsonb_array_elements(NEW.geojson->'features') AS arr(f)) s); EXCEPTION WHEN OTHERS THEN NEW.geom := NULL; END;
    ELSE NEW.geom := NULL;
    END IF;
  END;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_documents_validate_and_fill_geom ON documents;
CREATE TRIGGER trg_documents_validate_and_fill_geom BEFORE INSERT OR UPDATE ON documents FOR EACH ROW EXECUTE FUNCTION trg_documents_validate_and_fill_geom();

-- ========== Внутренние: get_or_create и очистка ==========
CREATE OR REPLACE FUNCTION _internal_get_or_create_author(p_name TEXT) RETURNS INTEGER LANGUAGE plpgsql AS $$
DECLARE v TEXT := btrim(p_name);
BEGIN
  IF v IS NULL OR v = '' THEN RETURN NULL; END IF;
  INSERT INTO authors (full_name) VALUES (v) ON CONFLICT (full_name) DO NOTHING;
  RETURN (SELECT id FROM authors WHERE full_name = v);
END; $$;

CREATE OR REPLACE FUNCTION _internal_get_or_create_tag(p_name TEXT) RETURNS INTEGER LANGUAGE plpgsql AS $$
DECLARE v TEXT := btrim(p_name);
BEGIN
  IF v IS NULL OR v = '' THEN RETURN NULL; END IF;
  INSERT INTO tags (name) VALUES (v) ON CONFLICT (name) DO NOTHING;
  RETURN (SELECT id FROM tags WHERE name = v);
END; $$;

CREATE OR REPLACE FUNCTION _internal_attach_tag_to_document(p_document_id INT, p_tag_name TEXT) RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE tid INT;
BEGIN
  IF p_tag_name IS NULL THEN RETURN; END IF;
  tid := _internal_get_or_create_tag(p_tag_name);
  INSERT INTO document_tags (document_id, tag_id) VALUES (p_document_id, tid) ON CONFLICT DO NOTHING;
END; $$;

CREATE OR REPLACE FUNCTION _internal_cleanup_unused_tags() RETURNS INTEGER
LANGUAGE plpgsql AS $$
DECLARE
  v_deleted INT := 0;
BEGIN
  DELETE FROM tags t
  WHERE NOT EXISTS (
    SELECT 1 FROM document_tags dt WHERE dt.tag_id = t.id
  );

  GET DIAGNOSTICS v_deleted := ROW_COUNT;
  RETURN v_deleted;
END;
$$;

CREATE OR REPLACE FUNCTION _internal_cleanup_unused_authors() RETURNS INTEGER
LANGUAGE plpgsql AS $$
DECLARE
  v_deleted INT := 0;
BEGIN
  DELETE FROM authors a
  WHERE NOT EXISTS (
    SELECT 1 FROM documents d WHERE d.author_id = a.id
  );

  GET DIAGNOSTICS v_deleted := ROW_COUNT;
  RETURN v_deleted;
END;
$$;

-- ========== Логирование (маска password_hash/file_bytea) ==========
CREATE OR REPLACE FUNCTION fn_log_changes() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_user_id INTEGER;
  v_user_login TEXT;
  v_new JSONB;
  v_old JSONB;
  v_tg_op TEXT := TG_OP;
  v_session_of_user TEXT := session_of_user; -- встроенная переменная
BEGIN
  -- попытка определить пользователя из created_by/updated_by, иначе current_user
  IF TG_OP = 'INSERT' THEN v_user_id := COALESCE(NEW.updated_by, NEW.created_by);
  ELSIF TG_OP = 'UPDATE' THEN v_user_id := COALESCE(NEW.updated_by, NEW.created_by);
  ELSE v_user_id := COALESCE(OLD.updated_by, OLD.created_by); END IF;

  IF v_user_id IS NOT NULL THEN SELECT login INTO v_user_login FROM users WHERE id = v_user_id; ELSE v_user_login := current_user; END IF;

  IF TG_OP = 'INSERT' THEN
    v_new := to_jsonb(NEW) - 'password_hash' - 'file_bytea';
    INSERT INTO logs(action, table_name, record_id, user_id, user_login, tg_op, session_of_user, action_time, changes)
    VALUES ('create'::action_type, TG_TABLE_NAME, NEW.id, v_user_id, v_user_login, v_tg_op, v_session_of_user, now(), jsonb_build_object('new', v_new));
    RETURN NEW;
  ELSIF TG_OP = 'UPDATE' THEN
    v_old := to_jsonb(OLD) - 'password_hash' - 'file_bytea';
    v_new := to_jsonb(NEW) - 'password_hash' - 'file_bytea';
    INSERT INTO logs(action, table_name, record_id, user_id, user_login, tg_op, session_of_user, action_time, changes)
    VALUES ('update'::action_type, TG_TABLE_NAME, NEW.id, v_user_id, v_user_login, v_tg_op, v_session_of_user, now(), jsonb_build_object('old', v_old, 'new', v_new));
    RETURN NEW;
  ELSE
    v_old := to_jsonb(OLD) - 'password_hash' - 'file_bytea';
    INSERT INTO logs(action, table_name, record_id, user_id, user_login, tg_op, session_of_user, action_time, changes)
    VALUES ('delete'::action_type, TG_TABLE_NAME, OLD.id, v_user_id, v_user_login, v_tg_op, v_session_of_user, now(), jsonb_build_object('old', v_old));
    RETURN OLD;
  END IF;
END; $$;

-- Очистка после удаления документа
CREATE OR REPLACE FUNCTION trg_after_delete_document_cleanup() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  PERFORM _internal_cleanup_unused_tags();
  PERFORM _internal_cleanup_unused_authors();
  RETURN OLD;
END; $$;
DROP TRIGGER IF EXISTS trg_after_delete_document_cleanup ON documents;
CREATE TRIGGER trg_after_delete_document_cleanup AFTER DELETE ON documents FOR EACH ROW EXECUTE FUNCTION trg_after_delete_document_cleanup();

-- ========== SECURITY-FUNCTIONS (SECURITY DEFINER + search_path) ==========
CREATE OR REPLACE FUNCTION fn_register_user(
  p_login TEXT,
  p_password TEXT,
  p_full_name TEXT
)
RETURNS INTEGER
SECURITY DEFINER
SET search_path = public, pg_temp
LANGUAGE plpgsql AS $$
DECLARE
  newid INT;
  v_role_id SMALLINT;
BEGIN
  -- Получаем id роли 'user'
  SELECT id INTO v_role_id FROM roles WHERE name = 'user';
  IF v_role_id IS NULL THEN
    RAISE EXCEPTION 'Role "user" not found. Create roles first or insert role ''user''.';
  END IF;

  INSERT INTO users (login, password_hash, full_name, role_id, created_at)
  VALUES (p_login::citext, p_password, p_full_name, v_role_id, now())
  RETURNING id INTO newid;

  RETURN newid;
EXCEPTION
  WHEN unique_violation THEN
    RAISE EXCEPTION 'User with login % already exists', p_login;
END;
$$;

-- Авторизация: сравниваем хеши напрямую (передаёте уже захешированный пароль)
CREATE OR REPLACE FUNCTION fn_authorize_user(p_login TEXT, p_password TEXT)
RETURNS TABLE (id INT, login citext, full_name TEXT, role_name TEXT) SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
BEGIN
  RETURN QUERY
  SELECT u.id, u.login, u.full_name, r.name
  FROM users u JOIN roles r ON r.id = u.role_id
  WHERE u.login = p_login::citext AND u.password_hash = p_password;
END; $$;

-- Проверки прав
CREATE OR REPLACE FUNCTION is_user_admin(p_user_id INT) RETURNS BOOLEAN SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
DECLARE rn TEXT;
BEGIN
  IF p_user_id IS NULL THEN RETURN FALSE; END IF;
  SELECT r.name INTO rn FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = p_user_id;
  RETURN rn = 'administrator';
END; $$;

CREATE OR REPLACE FUNCTION _can_user_edit_document(p_user_id INT, p_document_id INT) RETURNS BOOLEAN SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
DECLARE v_role TEXT; v_creator INT; v_perm BOOLEAN;
BEGIN
  IF p_user_id IS NULL THEN RETURN FALSE; END IF;
  SELECT r.name, d.created_by INTO v_role, v_creator FROM users u JOIN roles r ON r.id = u.role_id LEFT JOIN documents d ON d.id = p_document_id WHERE u.id = p_user_id LIMIT 1;
  IF v_role = 'administrator' OR v_creator = p_user_id THEN RETURN TRUE; END IF;
  SELECT EXISTS (SELECT 1 FROM document_permissions WHERE document_id = p_document_id AND user_id = p_user_id AND can_edit) INTO v_perm;
  RETURN v_perm;
END; $$;

CREATE OR REPLACE FUNCTION _can_user_view_document(p_user_id INT, p_document_id INT) RETURNS BOOLEAN SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
DECLARE v_role TEXT; v_privacy privacy_type; v_creator INT; v_perm BOOLEAN;
BEGIN
  IF p_user_id IS NOT NULL THEN SELECT r.name INTO v_role FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = p_user_id; END IF;
  SELECT d.privacy, d.created_by INTO v_privacy, v_creator FROM documents d WHERE d.id = p_document_id;
  IF v_role = 'administrator' OR v_privacy = 'public' OR v_creator = p_user_id THEN RETURN TRUE; END IF;
  SELECT EXISTS (SELECT 1 FROM document_permissions WHERE document_id = p_document_id AND user_id = p_user_id AND can_view) INTO v_perm;
  RETURN v_perm;
END; $$;

-- ========== CRUD для документов ==========
CREATE OR REPLACE FUNCTION fn_add_document(
  p_user_id INT, p_title TEXT, p_document_date DATE,
  p_author_id INT DEFAULT NULL, p_author_name TEXT DEFAULT NULL,
  p_type_id INT DEFAULT NULL, p_file BYTEA DEFAULT NULL,
  p_geojson JSONB DEFAULT NULL, p_tags TEXT[] DEFAULT NULL,
  p_privacy TEXT DEFAULT 'public'
) RETURNS INT SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
DECLARE new_id INT; a_id INT := p_author_id; t TEXT;
BEGIN
  IF p_user_id IS NULL THEN RAISE EXCEPTION 'p_user_id is required'; END IF;
  IF a_id IS NULL AND p_author_name IS NOT NULL THEN a_id := _internal_get_or_create_author(p_author_name); END IF;
  INSERT INTO documents (title, privacy, created_at, created_by, document_date, author_id, type_id, file_bytea, geojson)
  VALUES (p_title, p_privacy::privacy_type, now(), p_user_id, p_document_date, a_id, p_type_id, p_file, p_geojson) RETURNING id INTO new_id;

  IF p_tags IS NOT NULL THEN
    FOREACH t IN ARRAY p_tags LOOP PERFORM _internal_attach_tag_to_document(new_id, t); END LOOP;
  END IF;
  RETURN new_id;
END; $$;

CREATE OR REPLACE FUNCTION fn_update_document(
  p_document_id INT, p_user_id INT, p_title TEXT, p_document_date DATE,
  p_author_id INT DEFAULT NULL, p_author_name TEXT DEFAULT NULL,
  p_type_id INT DEFAULT NULL, p_file BYTEA DEFAULT NULL,
  p_geojson JSONB DEFAULT NULL, p_tags TEXT[] DEFAULT NULL, p_privacy TEXT DEFAULT NULL
) RETURNS VOID SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
DECLARE a_id INT := p_author_id; t TEXT;
BEGIN
  IF p_user_id IS NULL THEN RAISE EXCEPTION 'p_user_id is required'; END IF;
  IF NOT _can_user_edit_document(p_user_id, p_document_id) THEN RAISE EXCEPTION 'No permission'; END IF;
  IF a_id IS NULL AND p_author_name IS NOT NULL THEN a_id := _internal_get_or_create_author(p_author_name); END IF;

  UPDATE documents SET
    title = p_title, document_date = p_document_date, author_id = a_id, type_id = p_type_id,
    file_bytea = p_file, geojson = p_geojson,
    privacy = COALESCE(p_privacy, privacy)::privacy_type,
    updated_at = now(), updated_by = p_user_id
  WHERE id = p_document_id;

  IF p_tags IS NOT NULL THEN
    DELETE FROM document_tags WHERE document_id = p_document_id;
    IF COALESCE(array_length(p_tags,1),0) > 0 THEN
      FOREACH t IN ARRAY p_tags LOOP PERFORM _internal_attach_tag_to_document(p_document_id, t); END LOOP;
    END IF;
  END IF;
END; $$;

CREATE OR REPLACE FUNCTION fn_delete_document(p_document_id INT, p_user_id INT) RETURNS VOID SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
BEGIN
  IF p_user_id IS NULL THEN RAISE EXCEPTION 'p_user_id is required'; END IF;
  IF NOT _can_user_edit_document(p_user_id, p_document_id) THEN RAISE EXCEPTION 'No permission'; END IF;
  DELETE FROM documents WHERE id = p_document_id;
END; $$;

-- ========== Управление правами (только админ) ==========
CREATE OR REPLACE FUNCTION fn_set_document_permission(p_document_id INT, p_user_id INT, p_target_user_id INT, p_can_view BOOLEAN, p_can_edit BOOLEAN)
RETURNS VOID SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
BEGIN
  IF NOT is_user_admin(p_user_id) THEN RAISE EXCEPTION 'Only administrator may set permissions'; END IF;
  INSERT INTO document_permissions (document_id, user_id, can_view, can_edit)
    VALUES (p_document_id, p_target_user_id, p_can_view, p_can_edit)
    ON CONFLICT (document_id,user_id) DO UPDATE SET can_view = EXCLUDED.can_view, can_edit = EXCLUDED.can_edit;
END; $$;

CREATE OR REPLACE FUNCTION fn_remove_document_permission(p_document_id INT, p_user_id INT, p_target_user_id INT)
RETURNS VOID SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
BEGIN
  IF NOT is_user_admin(p_user_id) THEN RAISE EXCEPTION 'Only administrator may remove permissions'; END IF;
  DELETE FROM document_permissions WHERE document_id = p_document_id AND user_id = p_target_user_id;
END; $$;

-- ========== Поиск документов по тегу с проверкой прав (упрощённый экспорт) ==========
CREATE OR REPLACE FUNCTION fn_get_documents_by_tag_secure(
  p_tag_name TEXT, p_author_name TEXT DEFAULT NULL, p_type_name TEXT DEFAULT NULL,
  p_date_from DATE DEFAULT NULL, p_date_to DATE DEFAULT NULL, p_requester_id INT DEFAULT NULL
) RETURNS TABLE (
  doc_id INT, title TEXT, privacy privacy_type, created_at TIMESTAMPTZ, created_by INT,
  created_by_login TEXT, created_by_full_name TEXT, updated_at TIMESTAMPTZ, updated_by INT,
  updated_by_login TEXT, updated_by_full_name TEXT, document_date DATE, author_id INT, author_name TEXT,
  type_id INT, type_name TEXT, tags TEXT[], viewers INT[], editors INT[], can_requester_edit BOOLEAN, geom geometry(Geometry,4326)
) SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
DECLARE v_role TEXT; v_tag TEXT := btrim(p_tag_name); v_uid INT := p_requester_id;
BEGIN
  IF v_tag IS NULL OR v_tag = '' THEN RAISE EXCEPTION 'p_tag_name must be provided'; END IF;
  IF v_uid IS NOT NULL THEN SELECT r.name INTO v_role FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = v_uid; END IF;

  RETURN QUERY WITH tt AS (SELECT id FROM tags WHERE lower(name)=lower(v_tag))
  SELECT d.id, d.title, d.privacy, d.created_at, d.created_by, cu.login, cu.full_name,
        d.updated_at, d.updated_by, ud.login, ud.full_name, d.document_date,
        d.author_id, a.full_name, d.type_id, t.name,
        COALESCE((SELECT array_agg(DISTINCT tg ORDER BY tg) FROM (SELECT t2.name AS tg FROM document_tags dt2 JOIN tags t2 ON t2.id = dt2.tag_id WHERE dt2.document_id = d.id) s), ARRAY[]::text[]),
        COALESCE((SELECT array_agg(user_id) FROM document_permissions WHERE document_id = d.id AND can_view), ARRAY[]::int[]),
        COALESCE((SELECT array_agg(user_id) FROM document_permissions WHERE document_id = d.id AND can_edit), ARRAY[]::int[]),
        (CASE WHEN v_role='administrator' OR (v_uid IS NOT NULL AND d.created_by = v_uid) OR (v_uid IS NOT NULL AND EXISTS(SELECT 1 FROM document_permissions dp WHERE dp.document_id=d.id AND dp.user_id=v_uid AND dp.can_edit)) THEN TRUE ELSE FALSE END),
        d.geom
  FROM documents d
  JOIN document_tags dt ON dt.document_id = d.id
  JOIN tt ON tt.id = dt.tag_id
  LEFT JOIN authors a ON a.id = d.author_id
  LEFT JOIN document_types t ON t.id = d.type_id
  LEFT JOIN users cu ON cu.id = d.created_by
  LEFT JOIN users ud ON ud.id = d.updated_by
  WHERE (v_role='administrator' OR d.privacy='public' OR (v_uid IS NOT NULL AND d.created_by = v_uid) OR (v_uid IS NOT NULL AND EXISTS (SELECT 1 FROM document_permissions dp WHERE dp.document_id = d.id AND dp.user_id = v_uid AND dp.can_view)))
    AND (p_author_name IS NULL OR EXISTS (SELECT 1 FROM authors a2 WHERE a2.id = d.author_id AND lower(a2.full_name)=lower(btrim(p_author_name))))
    AND (p_type_name IS NULL OR EXISTS (SELECT 1 FROM document_types t2 WHERE t2.id = d.type_id AND lower(t2.name)=lower(btrim(p_type_name))))
    AND (p_date_from IS NULL OR d.document_date >= p_date_from)
    AND (p_date_to   IS NULL OR d.document_date <= p_date_to)
  ORDER BY d.created_at DESC;
END; $$;

-- ========== Получение одного документа с проверкой прав ==========
CREATE OR REPLACE FUNCTION fn_get_document_secure(p_document_id INT, p_requester_id INT DEFAULT NULL)
RETURNS TABLE (id INT, title TEXT, privacy privacy_type, created_at TIMESTAMPTZ, created_by INT, updated_at TIMESTAMPTZ, updated_by INT, document_date DATE, author_id INT, type_id INT, geojson JSONB)
SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
DECLARE v_role TEXT; v_uid INT := p_requester_id; allowed BOOLEAN;
BEGIN
  IF v_uid IS NOT NULL THEN SELECT r.name INTO v_role FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = v_uid; END IF;
  IF v_role='administrator' THEN RETURN QUERY SELECT id,title,privacy,created_at,created_by,updated_at,updated_by,document_date,author_id,type_id,geojson FROM documents WHERE id = p_document_id; RETURN; END IF;

  SELECT EXISTS (
    SELECT 1 FROM documents d
    WHERE d.id = p_document_id AND (d.privacy='public' OR (v_uid IS NOT NULL AND d.created_by = v_uid) OR (v_uid IS NOT NULL AND EXISTS (SELECT 1 FROM document_permissions dp WHERE dp.document_id=d.id AND dp.user_id=v_uid AND dp.can_view)))
  ) INTO allowed;

  IF NOT allowed THEN RAISE EXCEPTION 'User % has no permission to view document %', v_uid, p_document_id; END IF;
  RETURN QUERY SELECT id,title,privacy,created_at,created_by,updated_at,updated_by,document_date,author_id,type_id,geojson FROM documents WHERE id = p_document_id;
END; $$;

-- ========== Логи: получение (только админ) ==========
CREATE OR REPLACE FUNCTION fn_get_logs_by_user(p_requester_id INT, p_user_id_target INT, p_start TIMESTAMPTZ DEFAULT NULL, p_end TIMESTAMPTZ DEFAULT NULL)
RETURNS SETOF logs SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
BEGIN
  IF NOT is_user_admin(p_requester_id) THEN RAISE EXCEPTION 'Only administrator can access logs'; END IF;
  RETURN QUERY SELECT * FROM logs WHERE user_id = p_user_id_target AND (p_start IS NULL OR action_time >= p_start) AND (p_end IS NULL OR action_time <= p_end) ORDER BY action_time DESC;
END; $$;

CREATE OR REPLACE FUNCTION fn_get_logs_by_table(p_requester_id INT, p_table_name TEXT, p_start TIMESTAMPTZ DEFAULT NULL, p_end TIMESTAMPTZ DEFAULT NULL) RETURNS SETOF logs
SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
BEGIN
  IF NOT is_user_admin(p_requester_id) THEN RAISE EXCEPTION 'Only administrator can access logs'; END IF;
  RETURN QUERY SELECT * FROM logs WHERE table_name = p_table_name AND (p_start IS NULL OR action_time >= p_start) AND (p_end IS NULL OR action_time <= p_end) ORDER BY action_time DESC;
END; $$;

CREATE OR REPLACE FUNCTION fn_get_logs_by_date(p_requester_id INT, p_start TIMESTAMPTZ, p_end TIMESTAMPTZ) RETURNS SETOF logs
SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
BEGIN
  IF NOT is_user_admin(p_requester_id) THEN RAISE EXCEPTION 'Only administrator can access logs'; END IF;
  RETURN QUERY SELECT * FROM logs WHERE action_time >= p_start AND action_time <= p_end ORDER BY action_time DESC;
END; $$;

-- ========== Периодическая очистка ==========
CREATE OR REPLACE FUNCTION fn_periodic_cleanup() RETURNS JSONB
SECURITY DEFINER
SET search_path = public, pg_temp
LANGUAGE plpgsql AS $$
DECLARE
  v_start TIMESTAMPTZ := now();
  v_tags_deleted INT := 0;
  v_authors_deleted INT := 0;
  v_result JSONB;
BEGIN
  BEGIN
    DELETE FROM tags t
    WHERE NOT EXISTS (SELECT 1 FROM document_tags dt WHERE dt.tag_id = t.id);
    GET DIAGNOSTICS v_tags_deleted := ROW_COUNT;
  EXCEPTION WHEN OTHERS THEN
    v_tags_deleted := -1;
  END;

  BEGIN
    DELETE FROM authors a
    WHERE NOT EXISTS (SELECT 1 FROM documents d WHERE d.author_id = a.id);
    GET DIAGNOSTICS v_authors_deleted := ROW_COUNT;
  EXCEPTION WHEN OTHERS THEN
    v_authors_deleted := -1;
  END;

  v_result := jsonb_build_object(
    'started_at', to_jsonb(v_start),
    'finished_at', to_jsonb(now()),
    'tags_deleted', to_jsonb(v_tags_deleted),
    'authors_deleted', to_jsonb(v_authors_deleted)
  );

  RETURN v_result;
END;
$$;