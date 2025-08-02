-- Авторизация
CREATE TABLE users
(
    user_id       serial      PRIMARY KEY unique,
    username      varchar(255) not null unique,
    password_hash varchar(255) not null,
    user_type INT not null
);

-- Создание таблицы для данных о лаборантах
CREATE TABLE lab_technicians (
    lab_technician_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id),
    full_name VARCHAR(255) NOT NULL,
    last_login TIMESTAMP,
    services_provided JSON
);

-- Создание таблицы для бухгалтеров
CREATE TABLE accountants (
    accountant_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id),
    full_name VARCHAR(255) NOT NULL,
    last_login TIMESTAMP,
    invoices JSON
);

-- Создание таблицы для администраторов
CREATE TABLE administrators (
    administrator_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id)
);


-- Создание таблицы для услуг лаборатории
CREATE TABLE lab_services (
    service_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    cost DECIMAL(10, 2) NOT NULL,
    service_code VARCHAR(20) NOT NULL,
    execution_time INT NOT NULL,
    average_deviation DECIMAL(5, 2)
);

-- Создание таблицы для данных о страховых компаниях
CREATE TABLE insurance_companies (
    insurance_company_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    inn VARCHAR(20),
    bank_account VARCHAR(30),
    bik VARCHAR(20)
);

-- Создание таблицы для данных пациентов
CREATE TABLE patients (
    patient_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id),
    full_name VARCHAR(255) NOT NULL,
    date_of_birth DATE NOT NULL,
    passport_serial_number VARCHAR(20) NOT NULL,
    phone VARCHAR(15),
    email VARCHAR(255),
    insurance_number VARCHAR(20),
    insurance_type VARCHAR(50),
    insurance_company INT REFERENCES insurance_companies(insurance_company_id) NOT NULL
);


-- Создание таблицы для заказов
CREATE TABLE orders (
    order_id SERIAL PRIMARY KEY,
    creation_date DATE NOT NULL,
    patient_id INT REFERENCES patients(patient_id) NOT NULL,
    status_order VARCHAR(50),
    execution_time_in_days INT
);

-- Создание таблицы для оказанных услуг
CREATE TABLE provided_services (
    provided_service_id SERIAL PRIMARY KEY,
    service_id INT REFERENCES lab_services(service_id) NOT NULL,
    order_id INT REFERENCES orders(order_id) NOT NULL,
    execution_date TIMESTAMP,
    performer VARCHAR(255)
);

-- Создание таблицы для данных о работе анализаторов
CREATE TABLE analyzers (
    analyzer_id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(order_id) NOT NULL,
    arrival_date_time TIMESTAMP,
    completion_date_time TIMESTAMP,
    execution_time_in_seconds INT
);

-- Ограничение на отправку данных в архив
CREATE TABLE archived_data (
    data_id SERIAL PRIMARY KEY,
    table_name VARCHAR(50) NOT NULL,
    record_id INT NOT NULL,
    archive_date TIMESTAMP NOT NULL
);

CREATE TABLE user_sessions (
    session_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    session_duration INTERVAL,
    session_status VARCHAR(50)
);

CREATE TABLE failed_login_attempts (
    attempt_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id),
    attempt_time TIMESTAMP NOT NULL,
    ip_address VARCHAR(50),
    captcha_required BOOLEAN,
    captcha_text VARCHAR(4),
    blocked_for_interval INTERVAL
);

-- История Авторизации
CREATE TABLE login_history (
    login_history_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id) NOT NULL,
    login_time TIMESTAMP NOT NULL,
    success BOOLEAN NOT NULL
);

-- Ограничение на услуги в заказе
CREATE OR REPLACE FUNCTION prevent_order_archiving()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.table_name = 'orders' AND NOT EXISTS (
        SELECT 1 FROM provided_services WHERE provided_services.order_id = NEW.record_id
    ) THEN
        RAISE EXCEPTION 'Cannot archive order without provided services';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_order_archiving_trigger
BEFORE INSERT ON archived_data
FOR EACH ROW
EXECUTE FUNCTION prevent_order_archiving();