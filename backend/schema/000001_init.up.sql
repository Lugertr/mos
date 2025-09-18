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

-- === Справочники и пользователи/роли ===
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

CREATE TABLE IF NOT EXISTS document_types (id SERIAL PRIMARY KEY, name citext NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS tags (id SERIAL PRIMARY KEY, name citext NOT NULL UNIQUE);

-- === Документы и связанные сущности ===
CREATE TABLE IF NOT EXISTS documents (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  privacy privacy_type NOT NULL DEFAULT 'public',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by INTEGER REFERENCES users(id),
  updated_at TIMESTAMPTZ,
  updated_by INTEGER REFERENCES users(id),
  document_date DATE,
  author citext, -- теперь хранится имя автора как текст
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

-- === Индексы (ключевые) ===
CREATE INDEX IF NOT EXISTS documents_title_idx         ON documents (lower(title));
CREATE INDEX IF NOT EXISTS documents_title_trgm_idx    ON documents USING gin (lower(title) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS documents_created_at_idx    ON documents (created_at);
CREATE INDEX IF NOT EXISTS documents_document_date_idx ON documents (document_date);
CREATE INDEX IF NOT EXISTS documents_geom_gist         ON documents USING GIST (geom);
CREATE INDEX IF NOT EXISTS tags_name_lower_idx         ON tags (lower(name));
CREATE INDEX IF NOT EXISTS doc_types_name_idx          ON document_types (lower(name));
CREATE INDEX IF NOT EXISTS document_tags_tag_idx       ON document_tags (tag_id);
CREATE INDEX IF NOT EXISTS document_permissions_user_idx ON document_permissions (user_id);
CREATE INDEX IF NOT EXISTS logs_action_time_idx        ON logs (action_time);
CREATE INDEX IF NOT EXISTS logs_user_id_idx            ON logs (user_id);
-- полезный индекс по privacy для быстрых селектов публичных
CREATE INDEX IF NOT EXISTS documents_privacy_idx ON documents (privacy);

-- индекс для поиска по author
CREATE INDEX IF NOT EXISTS documents_author_lower_idx ON documents (lower(author));

-- === GEOJSON: валидация и заполнение geom (более устойчиво, с попыткой MakeValid) ===
CREATE OR REPLACE FUNCTION fn_validate_geojson(p_geojson JSONB) RETURNS BOOLEAN
LANGUAGE plpgsql STABLE AS $$
DECLARE g geometry;
begin
  IF p_geojson IS NULL THEN RETURN TRUE; END IF;
  CASE lower(coalesce(p_geojson->>'type',''))
    WHEN '' THEN RETURN FALSE;
    WHEN 'point','linestring','polygon','multipoint','multilinestring','multipolygon','geometrycollection' THEN
      BEGIN
        g := ST_SetSRID(ST_GeomFromGeoJSON(p_geojson::text),4326);
        IF ST_IsValid(g) THEN RETURN TRUE; END IF;
        -- пробуем сделать валидно
        IF ST_IsValid(ST_MakeValid(g)) THEN RETURN TRUE; ELSE RETURN FALSE; END IF;
      EXCEPTION WHEN OTHERS THEN RETURN FALSE; END;
    WHEN 'feature' THEN
      IF p_geojson->'geometry' IS NULL THEN RETURN FALSE; END IF;
      BEGIN
        g := ST_SetSRID(ST_GeomFromGeoJSON((p_geojson->'geometry')::text),4326);
        IF ST_IsValid(g) THEN RETURN TRUE; END IF;
        IF ST_IsValid(ST_MakeValid(g)) THEN RETURN TRUE; ELSE RETURN FALSE; END IF;
      EXCEPTION WHEN OTHERS THEN RETURN FALSE; END;
    WHEN 'featurecollection' THEN
      IF p_geojson->'features' IS NULL OR jsonb_typeof(p_geojson->'features') <> 'array' THEN RETURN FALSE; END IF;
      BEGIN
        RETURN ST_IsValid(ST_Collect(ARRAY(
          SELECT ST_SetSRID(ST_GeomFromGeoJSON((f->'geometry')::text),4326)
          FROM jsonb_array_elements(p_geojson->'features') AS arr(f)
        )));
      EXCEPTION WHEN OTHERS THEN
        -- пытаемся проверить с MakeValid
        BEGIN
          RETURN ST_IsValid(ST_MakeValid(ST_Collect(ARRAY(
            SELECT ST_SetSRID(ST_GeomFromGeoJSON((f->'geometry')::text),4326)
            FROM jsonb_array_elements(p_geojson->'features') AS arr(f)
          ))));
        EXCEPTION WHEN OTHERS THEN RETURN FALSE; END;
      END;
    ELSE RETURN FALSE;
  END CASE;
END;
$$;

CREATE OR REPLACE FUNCTION trg_documents_validate_and_fill_geom() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE g geometry;
BEGIN
  IF NEW.geojson IS NULL THEN NEW.geom := NULL; RETURN NEW; END IF;
  IF NOT fn_validate_geojson(NEW.geojson) THEN RAISE EXCEPTION 'Invalid GeoJSON for document %', COALESCE(NEW.title,'unknown'); END IF;
  BEGIN
    g := ST_SetSRID(ST_GeomFromGeoJSON(NEW.geojson::text),4326);
    IF NOT ST_IsValid(g) THEN
      g := ST_MakeValid(g);
    END IF;
    NEW.geom := g;
  EXCEPTION WHEN OTHERS THEN
    -- При ошибке пробуем разные варианты (feature / featurecollection)
    IF lower(coalesce(NEW.geojson->>'type',''))='feature' THEN
      BEGIN
        g := ST_SetSRID(ST_GeomFromGeoJSON((NEW.geojson->'geometry')::text),4326);
        IF NOT ST_IsValid(g) THEN g := ST_MakeValid(g); END IF;
        NEW.geom := g;
      EXCEPTION WHEN OTHERS THEN NEW.geom := NULL; END;
    ELSIF lower(coalesce(NEW.geojson->>'type',''))='featurecollection' THEN
      BEGIN
        NEW.geom := (SELECT ST_Collect(array_agg(g)) FROM (SELECT ST_SetSRID(ST_GeomFromGeoJSON((f->'geometry')::text),4326) AS g FROM jsonb_array_elements(NEW.geojson->'features') AS arr(f)) s);
        IF NEW.geom IS NOT NULL AND NOT ST_IsValid(NEW.geom) THEN NEW.geom := ST_MakeValid(NEW.geom); END IF;
      EXCEPTION WHEN OTHERS THEN NEW.geom := NULL; END;
    ELSE
      NEW.geom := NULL;
    END IF;
  END;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_documents_validate_and_fill_geom ON documents;
CREATE TRIGGER trg_documents_validate_and_fill_geom BEFORE INSERT OR UPDATE ON documents FOR EACH ROW EXECUTE FUNCTION trg_documents_validate_and_fill_geom();

CREATE OR REPLACE FUNCTION _internal_get_or_create_tag(p_name TEXT) RETURNS INTEGER LANGUAGE plpgsql AS $$
DECLARE v TEXT := btrim(p_name); r INT;
BEGIN
  IF v IS NULL OR v = '' THEN RETURN NULL; END IF;
  INSERT INTO tags (name) VALUES (v)
    ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO r;
  RETURN r;
END; $$;

CREATE OR REPLACE FUNCTION _internal_attach_tag_to_document(p_document_id INT, p_tag_name TEXT) RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE tid INT;
BEGIN
  IF p_tag_name IS NULL OR btrim(p_tag_name) = '' THEN RETURN; END IF;

  tid := _internal_get_or_create_tag(p_tag_name);
  IF tid IS NULL THEN RETURN; END IF;

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

-- === Логирование (маска password_hash/file_bytea). Теперь использует current_setting с безопасным флагом ===
CREATE OR REPLACE FUNCTION fn_log_changes() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
  v_user_id INTEGER;
  v_user_login TEXT;
  v_new JSONB;
  v_old JSONB;
  v_tg_op TEXT := TG_OP;
  v_session_of_user TEXT := current_setting('app.session_of_user', true);
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

-- Навесим триггер логирования на documents (и при желании на другие таблицы)
DROP TRIGGER IF EXISTS trg_log_changes_documents ON documents;
CREATE TRIGGER trg_log_changes_documents AFTER INSERT OR UPDATE OR DELETE ON documents
  FOR EACH ROW EXECUTE FUNCTION fn_log_changes();

CREATE OR REPLACE FUNCTION trg_after_delete_document_cleanup() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  PERFORM _internal_cleanup_unused_tags();
  RETURN OLD;
END; $$;
DROP TRIGGER IF EXISTS trg_after_delete_document_cleanup ON documents;
CREATE TRIGGER trg_after_delete_document_cleanup AFTER DELETE ON documents FOR EACH ROW EXECUTE FUNCTION trg_after_delete_document_cleanup();

-- Регистрация
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
  v_login citext;
BEGIN
  -- Валидация входных данных
  IF p_login IS NULL OR btrim(p_login) = '' THEN
    RAISE EXCEPTION 'p_login is required and must not be blank';
  END IF;
  IF p_password IS NULL OR btrim(p_password) = '' THEN
    RAISE EXCEPTION 'p_password is required and must not be blank';
  END IF;

  v_login := btrim(p_login)::citext;

  -- Получаем id роли 'user'
  SELECT id INTO v_role_id FROM roles WHERE name = 'user';
  IF v_role_id IS NULL THEN
    RAISE EXCEPTION 'Role "user" not found. Create roles first or insert role ''user''.';
  END IF;

  INSERT INTO users (login, password_hash, full_name, role_id, created_at)
  VALUES (v_login, p_password, p_full_name, v_role_id, now())
  RETURNING id INTO newid;

  RETURN newid;
EXCEPTION
  WHEN unique_violation THEN
    RAISE EXCEPTION 'User with login % already exists', v_login;
END;
$$;

-- Авторизация
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

CREATE OR REPLACE FUNCTION _can_user_view_document(p_user_id INT, p_document_id INT) RETURNS BOOLEAN
SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
DECLARE v_role TEXT; v_privacy privacy_type; v_creator INT; v_perm BOOLEAN;
BEGIN
  IF p_user_id IS NOT NULL THEN SELECT r.name INTO v_role FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = p_user_id; END IF;
  SELECT d.privacy, d.created_by INTO v_privacy, v_creator FROM documents d WHERE d.id = p_document_id;
  IF v_role = 'administrator' OR v_privacy = 'public'::privacy_type OR v_creator = p_user_id THEN RETURN TRUE; END IF;
  SELECT EXISTS (SELECT 1 FROM document_permissions WHERE document_id = p_document_id AND user_id = p_user_id AND can_view) INTO v_perm;
  RETURN v_perm;
END; $$;

-- Создание документа
CREATE OR REPLACE FUNCTION fn_add_document(
  p_user_id INT,
  p_title TEXT,
  p_document_date DATE,
  p_author TEXT,     
  p_type_id INT,
  p_file BYTEA,
  p_geojson JSONB,
  p_tags TEXT[],
  p_privacy TEXT
) RETURNS INT
SECURITY DEFINER
SET search_path = public, pg_temp
LANGUAGE plpgsql AS $$
DECLARE
  new_id INT;
  a_name citext := NULL;
  t TEXT;
  v_privacy_lower TEXT;
BEGIN
  -- Обязательные проверки
  IF p_user_id IS NULL THEN RAISE EXCEPTION 'p_user_id is required'; END IF;
  IF p_title IS NULL OR btrim(p_title) = '' THEN RAISE EXCEPTION 'p_title is required'; END IF;
  IF p_privacy IS NULL THEN RAISE EXCEPTION 'p_privacy must be provided'; END IF;

  v_privacy_lower := lower(btrim(p_privacy));
  IF v_privacy_lower NOT IN ('public','private') THEN
    RAISE EXCEPTION 'p_privacy must be one of public/private';
  END IF;

  IF p_author IS NOT NULL AND btrim(p_author) <> '' THEN
    a_name := p_author::citext;
  END IF;

  -- Вставка документа (atomic в рамках функции)
  INSERT INTO documents (title, privacy, created_at, created_by, document_date, author, type_id, file_bytea, geojson)
  VALUES (p_title, v_privacy_lower::privacy_type, now(), p_user_id, p_document_date, a_name, p_type_id, p_file, p_geojson)
  RETURNING id INTO new_id;

  -- Теги: убираем дубликаты, создаём и привязываем
  IF p_tags IS NOT NULL THEN
    FOR t IN SELECT DISTINCT btrim(tag) FROM unnest(p_tags) tag WHERE btrim(tag) <> '' LOOP
      PERFORM _internal_attach_tag_to_document(new_id, t);
    END LOOP;
  END IF;

  RETURN new_id;
END;
$$;

CREATE OR REPLACE FUNCTION fn_update_document(
  p_document_id INT,
  p_user_id INT,
  p_title TEXT,
  p_document_date DATE,
  p_author TEXT,       
  p_type_id INT,
  p_file BYTEA,
  p_geojson JSONB,
  p_tags TEXT[],
  p_privacy TEXT
) RETURNS VOID
SECURITY DEFINER
SET search_path = public, pg_temp
LANGUAGE plpgsql AS $$
DECLARE
  a_name citext := NULL;
  t TEXT;
  v_exists BOOLEAN;
  v_privacy_lower TEXT;
BEGIN
  -- Обязательные проверки
  IF p_user_id IS NULL THEN RAISE EXCEPTION 'p_user_id is required'; END IF;
  IF p_document_id IS NULL THEN RAISE EXCEPTION 'p_document_id is required'; END IF;
  IF p_title IS NULL OR btrim(p_title) = '' THEN RAISE EXCEPTION 'p_title is required'; END IF;
  IF p_privacy IS NULL THEN RAISE EXCEPTION 'p_privacy must be provided'; END IF;

  v_privacy_lower := lower(btrim(p_privacy));
  IF v_privacy_lower NOT IN ('public','private') THEN
    RAISE EXCEPTION 'p_privacy must be one of public/private';
  END IF;

  IF NOT _can_user_edit_document(p_user_id, p_document_id) THEN
    RAISE EXCEPTION 'No permission';
  END IF;

  -- проверим существование
  SELECT EXISTS (SELECT 1 FROM documents WHERE id = p_document_id) INTO v_exists;
  IF NOT v_exists THEN RAISE EXCEPTION 'Document % does not exist', p_document_id; END IF;

  IF p_author IS NOT NULL AND btrim(p_author) <> '' THEN
    a_name := p_author::citext;
  END IF;

  -- Обновляем запись
  UPDATE documents SET
    title = p_title,
    document_date = p_document_date,
    author = a_name,
    type_id = p_type_id,
    file_bytea = p_file,
    geojson = p_geojson,
    privacy = v_privacy_lower::privacy_type,
    updated_at = now(),
    updated_by = p_user_id
  WHERE id = p_document_id;

  -- Теги: если передан NULL — не трогаем; если передан массив (включая пустой) — заменяем
  IF p_tags IS NOT NULL THEN
    DELETE FROM document_tags WHERE document_id = p_document_id;
    FOR t IN SELECT DISTINCT btrim(tag) FROM unnest(p_tags) tag WHERE btrim(tag) <> '' LOOP
      PERFORM _internal_attach_tag_to_document(p_document_id, t);
    END LOOP;
  END IF;
END;
$$;

CREATE OR REPLACE FUNCTION fn_delete_document(p_document_id INT, p_user_id INT) RETURNS VOID SECURITY DEFINER SET search_path = public, pg_temp LANGUAGE plpgsql AS $$
BEGIN
  IF p_user_id IS NULL THEN RAISE EXCEPTION 'p_user_id is required'; END IF;
  IF NOT _can_user_edit_document(p_user_id, p_document_id) THEN RAISE EXCEPTION 'No permission'; END IF;
  DELETE FROM documents WHERE id = p_document_id;
END; $$;

-- === Управление правами (только админ) ===
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

-- === Получение документов для пользователя ===
CREATE OR REPLACE FUNCTION fn_get_documents_for_user(p_requester_id INT)
RETURNS TABLE (
  id INT,
  title TEXT,
  privacy privacy_type,
  updated_at TIMESTAMPTZ,
  document_date DATE,
  type_id INT,
  author citext,
  geojson JSONB,
  can_edit BOOLEAN,
  is_author BOOLEAN
)
SECURITY DEFINER
SET search_path = public, pg_temp
LANGUAGE plpgsql AS $$
DECLARE
  v_role TEXT;
  v_uid INT := p_requester_id;
BEGIN
  IF v_uid IS NOT NULL THEN
    SELECT r.name INTO v_role FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = v_uid;
  END IF;

  RETURN QUERY
  SELECT
    d.id,
    d.title,
    d.privacy,
    d.updated_at,
    d.document_date,
    d.type_id,
    d.author,
    d.geojson,
    (CASE WHEN v_role = 'administrator' THEN TRUE
          WHEN v_uid IS NOT NULL AND d.created_by = v_uid THEN TRUE
          WHEN v_uid IS NOT NULL AND EXISTS (SELECT 1 FROM document_permissions dp WHERE dp.document_id = d.id AND dp.user_id = v_uid AND dp.can_edit) THEN TRUE
          ELSE FALSE END) AS can_edit,
    (d.created_by IS NOT NULL AND v_uid IS NOT NULL AND d.created_by = v_uid) AS is_author
  FROM documents d
  WHERE
    (v_role = 'administrator')
    OR d.privacy = 'public'::privacy_type
    OR (v_uid IS NOT NULL AND d.created_by = v_uid)
    OR (v_uid IS NOT NULL AND EXISTS (SELECT 1 FROM document_permissions dp WHERE dp.document_id = d.id AND dp.user_id = v_uid AND (dp.can_view OR dp.can_edit)))
  ORDER BY d.created_at DESC;
END;
$$;

-- === Получение полного представления одного документа (с проверкой прав) ===
CREATE OR REPLACE FUNCTION fn_get_document_by_id(p_document_id INT, p_requester_id INT DEFAULT NULL)
RETURNS TABLE (
  id INT,
  title TEXT,
  privacy privacy_type,
  created_at TIMESTAMPTZ,
  created_by INT,
  updated_at TIMESTAMPTZ,
  updated_by INT,
  document_date DATE,
  author citext,
  type_id INT,
  file_bytea BYTEA,
  geojson JSONB,
  geom geometry(Geometry,4326),
  can_edit BOOLEAN
)
SECURITY DEFINER
SET search_path = public, pg_temp
LANGUAGE plpgsql AS $$
DECLARE
  v_role TEXT;
  v_uid INT := p_requester_id;
  allowed BOOLEAN;
BEGIN
  IF v_uid IS NOT NULL THEN
    SELECT r.name INTO v_role FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = v_uid;
  END IF;

  IF v_role = 'administrator' THEN
    RETURN QUERY
      SELECT d.id, d.title, d.privacy, d.created_at, d.created_by, d.updated_at, d.updated_by,
             d.document_date, d.author, d.type_id, d.file_bytea, d.geojson, d.geom,
             TRUE AS can_edit
      FROM documents d
      WHERE d.id = p_document_id;
    RETURN;
  END IF;

  SELECT EXISTS (
    SELECT 1 FROM documents d
    WHERE d.id = p_document_id
      AND (
        d.privacy = 'public'::privacy_type
        OR (v_uid IS NOT NULL AND d.created_by = v_uid)
        OR (v_uid IS NOT NULL AND EXISTS (SELECT 1 FROM document_permissions dp WHERE dp.document_id = d.id AND dp.user_id = v_uid AND dp.can_view))
      )
  ) INTO allowed;

  IF NOT allowed THEN
    RAISE EXCEPTION 'User % has no permission to view document %', v_uid, p_document_id;
  END IF;

  RETURN QUERY
    SELECT d.id, d.title, d.privacy, d.created_at, d.created_by, d.updated_at, d.updated_by,
           d.document_date, d.author, d.type_id, d.file_bytea, d.geojson, d.geom,
           (CASE WHEN v_role = 'administrator' THEN TRUE
                 WHEN v_uid IS NOT NULL AND d.created_by = v_uid THEN TRUE
                 WHEN v_uid IS NOT NULL AND EXISTS (SELECT 1 FROM document_permissions dp WHERE dp.document_id = d.id AND dp.user_id = v_uid AND dp.can_edit) THEN TRUE
                 ELSE FALSE END) AS can_edit
    FROM documents d
    WHERE d.id = p_document_id;
END;
$$;
-- === Периодическая очистка ===
CREATE OR REPLACE FUNCTION fn_periodic_cleanup() RETURNS JSONB
SECURITY DEFINER
SET search_path = public, pg_temp
LANGUAGE plpgsql AS $$
DECLARE
  v_start TIMESTAMPTZ := now();
  v_tags_deleted INT := 0;
  v_result JSONB;
BEGIN
  BEGIN
    DELETE FROM tags t
    WHERE NOT EXISTS (SELECT 1 FROM document_tags dt WHERE dt.tag_id = t.id);
    GET DIAGNOSTICS v_tags_deleted := ROW_COUNT;
  EXCEPTION WHEN OTHERS THEN
    v_tags_deleted := -1;
  END;

  v_result := jsonb_build_object(
    'started_at', to_jsonb(v_start),
    'finished_at', to_jsonb(now()),
    'tags_deleted', to_jsonb(v_tags_deleted)
  );

  RETURN v_result;
END;
$$;

-- ========== Подготовка: роли (на случай, если ещё нет) ==========
INSERT INTO roles (id,name)
SELECT * FROM (VALUES (1,'administrator'),(2,'user')) v(id,name)
ON CONFLICT DO NOTHING;

-- ========== Справочники: типы документов ==========
INSERT INTO document_types (id, name)
VALUES
  (1, 'report'),
  (2, 'map'),
  (3, 'note')
ON CONFLICT (name) DO NOTHING;

-- ========== Пользователи (явно указываем id чтобы удобно ссылаться) ==========
INSERT INTO users (id, role_id, login, password_hash, full_name, created_at)
VALUES
  (1, 1, 'admin', 'hash_admin_123', 'Admin User', now()),
  (2, 2, 'alice', 'hash_alice_123', 'Alice Example', now()),
  (3, 2, 'bob',   'hash_bob_123',   'Bob Example',   now())
ON CONFLICT (login) DO NOTHING;

-- ========= Необязательные стартовые тэги (можно не выполнять, теги будут созданы функцией) =========
INSERT INTO tags (name) VALUES
  ('park'), ('survey'), ('strategy'), ('confidential'),
  ('environment'), ('cleanup'), ('boundary'), ('city'), ('notes')
ON CONFLICT (name) DO NOTHING;

-- ========== Вставка 5 документов через fn_add_document (возвращает id) ==========

-- 1) Public point (created by Alice)
SELECT fn_add_document(
  2, -- p_user_id = alice
  'Survey of Central Park', -- title
  '2025-06-01'::date, -- document_date
  'Dr. Alice', -- author
  2, -- type_id = map
  NULL, -- file bytea
  NULL,
  ARRAY['park','survey'], -- tags
  'public' -- privacy
) AS new_document_id;


-- 2) Private report (created by Admin) — дадим доступ пользователю bob
WITH d AS (
  SELECT fn_add_document(
    1, -- admin
    'Confidential Strategy', -- title
    '2024-12-15'::date,
    'Chief Strategist',
    1, -- type_id = report
    NULL,
    NULL::jsonb, -- no geojson
    ARRAY['strategy','confidential'],
    'private'
  ) AS id
)
-- grant view permission for Bob (user id = 3) by calling admin-setter
SELECT fn_set_document_permission((SELECT id FROM d), 1, 3, true, false) FROM d;

-- 3) Public linestring (River cleanup by Bob)
SELECT fn_add_document(
  3, -- bob
  'River Cleanup 2024',
  '2024-08-20'::date,
  'Bob',
  1, -- report
  NULL,
  $${
    "type":"LineString",
    "coordinates":[
      [30.4500,50.4200],
      [30.4600,50.4250],
      [30.4700,50.4300]
    ]
  }$$::jsonb,
  ARRAY['environment','cleanup'],
  'public'
) AS new_document_id;

-- 4) Public polygon (City boundary) — geojson as Feature (polygon)
SELECT fn_add_document(
  1, -- admin creates
  'City Boundary',
  '2020-01-01'::date,
  'City Council',
  2, -- map
  NULL,
  $${
    "type":"Feature",
    "properties": {"name":"City limits"},
    "geometry": {
      "type":"Polygon",
      "coordinates":[
        [
          [30.40,50.40],
          [30.55,50.40],
          [30.55,50.55],
          [30.40,50.55],
          [30.40,50.40]
        ]
      ]
    }
  }$$::jsonb,
  ARRAY['boundary','city'],
  'public'
) AS new_document_id;

-- 5) Public notes (simple document without geojson)
SELECT fn_add_document(
  2, -- alice
  'Meeting Notes — 2025-01-10',
  '2025-01-10'::date,
  'Alice Example',
  3, -- note
  NULL,
  NULL::jsonb,
  ARRAY['notes'],
  'public'
) AS new_document_id;

-- ========== Проверка: вывести все документы и связанные тэги/создателей ==========
-- Основной быстрый просмотр
SELECT d.id, d.title, d.privacy, d.document_date, d.author, dt.name AS type_name, u.login AS created_by_login
FROM documents d
LEFT JOIN document_types dt ON dt.id = d.type_id
LEFT JOIN users u ON u.id = d.created_by
ORDER BY d.created_at DESC;

-- Тэги (каждая строка = документ + тэг)
SELECT d.id AS document_id, d.title, t.name AS tag
FROM documents d
JOIN document_tags dtg ON dtg.document_id = d.id
JOIN tags t ON t.id = dtg.tag_id
ORDER BY d.id, t.name;
