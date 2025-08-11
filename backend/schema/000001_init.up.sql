CREATE TABLE users
(
    id            serial       not null unique,
    name          varchar(255) not null,
    username      varchar(255) not null unique,
    password_hash varchar(255) not null,
    acc_status boolean
);

CREATE TABLE app_type_table 
( 
    app_type_id SERIAL    NOT NULL UNIQUE, 
    app_type_name TEXT NOT NULL 
);

CREATE TABLE app_table 
( 
    app_id SERIAL    NOT NULL UNIQUE, 
    rooms BIGINT NOT NULL, 
    app_type_id BIGINT NOT NULL, 
    app_status BIGINT NOT NULL,
    app_price DOUBLE PRECISION NOT NULL 
);

CREATE TABLE client_table 
( 
    client_id   SERIAL    NOT NULL UNIQUE, 
    client_name TEXT NOT NULL, 
    family_name TEXT NOT NULL, 
    surname TEXT NOT NULL, 
    passport TEXT NOT NULL, 
    gender TEXT NOT NULL,
    app_id SERIAL NOT NULL , 
    date_in TEXT NOT NULL, 
    date_out TEXT NOT NULL
);

CREATE TABLE service_type_table ( 
    service_type_id SERIAL    NOT NULL UNIQUE,  
    service_type_name TEXT NOT NULL, 
    price DOUBLE PRECISION NOT NULL 
);

CREATE TABLE service_table 
( 
    service_id SERIAL    NOT NULL UNIQUE, 
    client_id SERIAL NOT NULL , 
    service_type_id SERIAL NOT NULL, 
    days_count BIGINT NOT NULL 
);

CREATE TABLE old_client_table 
( 
    old_client_id   SERIAL    NOT NULL UNIQUE PRIMARY KEY, 
    old_client_name TEXT NOT NULL, 
    old_family_name TEXT NOT NULL, 
    old_surname TEXT NOT NULL, 
    old_passport TEXT NOT NULL, 
    old_gender TEXT NOT NULL,
    old_app_id SERIAL NOT NULL, 
    old_date_in TEXT NOT NULL, 
    old_date_out TEXT NOT NULL
);

CREATE TABLE old_services_table 
( 
    old_services_id SERIAL NOT NULL UNIQUE PRIMARY KEY,
    old_passport TEXT NOT NULL, 
    old_services TEXT NOT NULL
);

CREATE OR REPLACE FUNCTION validate_passport_format() 
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.passport ~ '^\d{4}\s\d{6}$' OR NEW.passport ~ '^[IVXA]+\d{6}$' THEN    
    RETURN NEW;
  END IF;
    RAISE EXCEPTION 'Invalid passport format: %', NEW.passport;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER client_passport_trigger 
BEFORE INSERT OR UPDATE ON client_table 
FOR EACH ROW 
EXECUTE FUNCTION validate_passport_format();

CREATE OR REPLACE FUNCTION get_client_services(val_client_id numeric) RETURNS TEXT AS $$
DECLARE
  services TEXT := '';
  row Record;
BEGIN
  FOR row IN SELECT st.service_type_id, st.service_type_name, s.days_count, st.price
            FROM service_table s
            JOIN service_type_table st ON s.service_type_id = st.service_type_id
            WHERE s.client_id = val_client_id
            ORDER BY s.service_type_id
  LOOP
    services := services || row.service_type_id || ' ' || row.service_type_name || ' ' || row.price || ', ';
  END LOOP;
  IF LENGTH(services) > 2 THEN
    services := SUBSTRING(services, 1, LENGTH(services) - 2);
  END IF;
  RETURN services;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION client_deletion_trigger_function()
  RETURNS TRIGGER AS $$
DECLARE 
  services text := '';
BEGIN
    INSERT INTO old_client_table (
    old_client_name,
    old_family_name,
    old_surname,
    old_passport,
    old_gender,
    old_app_id,
    old_date_in,
    old_date_out
  )
  VALUES (
    OLD.client_name,
    OLD.family_name,
    OLD.surname,
    OLD.passport,
    OLD.gender,
    OLD.app_id,
    OLD.date_in,
    OLD.date_out
  );

  SELECT * INTO services FROM get_client_services(OLD.client_id);
  DELETE FROM service_table WHERE client_id = OLD.client_id;
  
  IF services IS NOT NULL THEN 
    INSERT INTO old_services_table  (
      old_passport,
      old_services
    )
    VALUES (
      OLD.passport,
      services
    );
  END IF;
  RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER client_deletion_trigger
AFTER DELETE ON client_table
FOR EACH ROW
EXECUTE FUNCTION client_deletion_trigger_function();

CREATE OR REPLACE FUNCTION count_days(cur_client_id numeric) RETURNS INTEGER AS $$
DECLARE
  days INTEGER := 0;
  date_in DATE ;
  date_out DATE ;
  current_date DATE := CURRENT_DATE;
BEGIN
  SELECT client_table.date_in, client_table.date_out INTO date_in, date_out 
  FROM client_table WHERE client_table.client_id = cur_client_id;

  IF date_out < current_date THEN
    days := date_out - date_in;
  ELSE
    days := current_date - date_in;
  END IF;

  IF days > 0 THEN
    days := days;
  END IF;

  RETURN days;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_client_services(val_client_id numeric) 
RETURNS TEXT AS $$
DECLARE 
	services TEXT := ''; 
	row Record; 
BEGIN 
	FOR row IN SELECT st.service_type_id, st.service_type_name, s.days_count, st.price 
	FROM service_table s 
	JOIN service_type_table st 
	ON s.service_type_id = st.service_type_id 
	WHERE s.client_id = val_client_id 
	ORDER BY s.service_type_id 
	LOOP 
		services := services || row.service_type_id || ' ' || row.service_type_name || ' ' || row.price || ' ' || row.days_count || ', '; 
	END LOOP; 
	IF LENGTH(services) > 2 THEN 
		services := SUBSTRING(services, 1, LENGTH(services) - 2); 
	END IF; 
RETURN services; 
END; 
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION get_client_info(cur_client_id numeric)
RETURNS TABLE (
    name text,
    app_id bigint,
    rooms bigint,
    app_price double precision,
    service_total double precision,
    discount double precision,
    services text,
    cl_count_days bigint
) AS $$
BEGIN
    -- Запрос 1: формирование имени клиента
    SELECT CONCAT(client_table.family_name, ' ', LEFT(client_table.client_name, 1), '.', LEFT(client_table.surname, 1), '.') 
    INTO name
    FROM client_table 
    WHERE client_id = cur_client_id;

    -- Запрос 2: app_id для cur_client_id
    SELECT client_table.app_id
    INTO app_id
    FROM client_table 
    WHERE client_id = cur_client_id;

    -- Запрос 3: rooms для cur_client_id
    SELECT app_table.rooms
    INTO rooms
    FROM app_table 
    WHERE app_table.app_id = (SELECT client_table.app_id FROM client_table WHERE client_id = cur_client_id);

    -- Запрос 4: app_price для cur_client_id
    SELECT app_table.app_price
    INTO app_price
    FROM app_table 
    WHERE app_table.app_id = (SELECT client_table.app_id FROM client_table WHERE client_id = cur_client_id);

    -- Запрос 5: подсчет суммы услуг для cur_client_id
    SELECT SUM(service_table.days_count * service_type_table.price)
    INTO service_total
    FROM service_table 
    JOIN service_type_table ON service_table.service_type_id = service_type_table.service_type_id
    WHERE service_table.client_id = cur_client_id;
	
	IF service_total is null THEN
      service_total := 0;
    END IF;
    -- Запрос 6: проверка наличия пасспорта в таблице old_client_table и old_services_table
    SELECT CASE WHEN EXISTS(SELECT * FROM old_client_table WHERE old_passport = client_table.passport) 
                AND EXISTS(SELECT * FROM old_services_table WHERE old_passport = client_table.passport)
                THEN 0.1
                WHEN EXISTS(SELECT * FROM old_client_table WHERE old_passport = client_table.passport) 
                THEN 0.05
                ELSE 0
    END
    INTO discount
    FROM client_table 
    WHERE client_id = cur_client_id;

    -- Запрос 7: результат функции get_client_services(cur_client_id)
    SELECT * 
    INTO services
    FROM get_client_services(cur_client_id);
    IF services is null THEN
      services := '';
    END IF;

    -- Запрос 8: результат функции get_client_services(cur_client_id)
    SELECT * 
    INTO cl_count_days
    FROM count_days(cur_client_id);

    IF cl_count_days is null THEN
      cl_count_days := 0;
    END IF;

    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_service_types()
RETURNS TABLE (service_type_name TEXT) AS $$
BEGIN
    RETURN QUERY
    SELECT st.service_type_name
    FROM service_type_table st
    LEFT JOIN (
        SELECT service_type_id, COUNT(*) as count
        FROM service_table
        GROUP BY service_type_id
    ) s ON st.service_type_id = s.service_type_id
    WHERE s.service_type_id IS NULL OR s.count < 1;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_today_tomorrow()
RETURNS SETOF app_table AS $$
BEGIN
    RETURN QUERY SELECT * FROM app_table 
    WHERE app_status = 1 AND app_id IN 
        (SELECT app_id FROM client_table 
        WHERE to_char(to_date(date_out, 'YYYY-MM-DD'), 'YYYY-MM-DD') IN (to_char(CURRENT_DATE, 'YYYY-MM-DD'), to_char(CURRENT_DATE + INTERVAL '1 day', 'YYYY-MM-DD')));
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_app_table_on_insert() RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS(SELECT 1 FROM app_table WHERE app_id = NEW.app_id AND app_status = 1) THEN
        RAISE EXCEPTION 'Error: app_id already exists with app_status = 1';
    ELSEIF NOT EXISTS(SELECT 1 FROM app_table WHERE app_id = NEW.app_id AND app_status = 0) THEN
        INSERT INTO app_table (app_id, rooms, app_type_id, app_status, app_price) 
        VALUES (NEW.app_id, 0, 0, 1, 0);
    ELSE
        UPDATE app_table SET app_status = 1 WHERE app_id = NEW.app_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER insert_client_trigger
BEFORE INSERT ON client_table
FOR EACH ROW
EXECUTE FUNCTION update_app_table_on_insert();

CREATE OR REPLACE FUNCTION update_app_table_on_update() RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS(SELECT 1 FROM app_table WHERE app_id = OLD.app_id AND app_status = 1) THEN
        UPDATE app_table SET app_status = 0 WHERE app_id = OLD.app_id;
    END IF;
    IF NOT EXISTS(SELECT 1 FROM app_table WHERE app_id = NEW.app_id AND app_status = 0) THEN
        INSERT INTO app_table (app_id, rooms, app_type_id, app_status, app_price) 
        VALUES (NEW.app_id, 0, 0, 1, 0);
    ELSEIF EXISTS(SELECT 1 FROM app_table WHERE app_id = NEW.app_id AND app_status = 1) THEN
        RAISE EXCEPTION 'Error: app_id already exists with app_status = 1';
    ELSE
        UPDATE app_table SET app_status = 1 WHERE app_id = NEW.app_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_client_trigger
BEFORE UPDATE ON client_table
FOR EACH ROW
EXECUTE FUNCTION update_app_table_on_update();

CREATE OR REPLACE FUNCTION update_app_table_on_delete() RETURNS TRIGGER AS $$
BEGIN
    UPDATE app_table SET app_status = 0 WHERE app_id = OLD.app_id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER delete_client_trigger
BEFORE DELETE ON client_table
FOR EACH ROW
EXECUTE FUNCTION update_app_table_on_delete();

CREATE OR REPLACE FUNCTION check_date() 
RETURNS TRIGGER 
AS $$ 
BEGIN 
	IF NEW.date_in >= NEW.date_out THEN 
		RAISE EXCEPTION 'Date_in must be before date_out'; 
	END IF; RETURN NEW; END; 
$$ LANGUAGE plpgsql;

CREATE TRIGGER date_check_trigger 
BEFORE INSERT OR UPDATE ON client_table 
FOR EACH ROW 
EXECUTE FUNCTION check_date();

CREATE OR REPLACE FUNCTION check_service_type_exists() RETURNS TRIGGER AS $$
BEGIN
  IF EXISTS (SELECT 1 FROM service_table WHERE client_id = NEW.client_id AND service_type_id = NEW.service_type_id) THEN
    RAISE EXCEPTION 'Service type already exists for this client';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER service_table_check_service_type_exists
BEFORE INSERT ON service_table
FOR EACH ROW
EXECUTE FUNCTION check_service_type_exists();



INSERT INTO users (name, username, password_hash,acc_status) VALUES ('admin', 'admin', '686a7172686a7177313234363137616a6668616a738c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918','true');

INSERT INTO app_type_table (app_type_name) VALUES ('Обычная'), ('Люкс'), ('ПолуЛюкс');

INSERT INTO app_table (rooms, app_type_id, app_status, app_price) VALUES (1, 1, 0, 5000), (2, 2, 0, 7500), (4, 3, 0, 12000);

INSERT INTO client_table (client_name, family_name, surname, passport, gender, app_id, date_in, date_out) VALUES ('John', 'Doe', 'Smith', '1234 567890', 'Male', 1, '2021-10-01', '2021-10-05');
INSERT INTO client_table (client_name, family_name, surname, passport, gender, app_id, date_in, date_out) VALUES ('JohnQ', 'DoeQ', 'SmithQ', 'IVA685635', 'Male', 2, '2023-05-19', '2023-05-22');

INSERT INTO service_type_table (service_type_name, price) VALUES ('Breakfast', 10.00), ('Lunch', 15.00), ('Dinner', 20.00);

INSERT INTO service_table (client_id, service_type_id, days_count) VALUES (1, 1, 4);