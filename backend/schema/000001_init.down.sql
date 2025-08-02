-- Удаление триггеров
DROP TRIGGER IF EXISTS prevent_order_archiving_trigger ON archived_data;

-- Удаление таблиц
DROP TABLE IF EXISTS provided_services;
DROP TABLE IF EXISTS analyzers;
DROP TABLE IF EXISTS lab_technicians;
DROP TABLE IF EXISTS accountants;
DROP TABLE IF EXISTS administrators;
DROP TABLE IF EXISTS archived_data;
DROP TABLE IF EXISTS failed_login_attempts;
DROP TABLE IF EXISTS login_history;
DROP TABLE IF EXISTS lab_services;
DROP TABLE IF EXISTS patients;
DROP TABLE IF EXISTS insurance_companies;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS users;


-- Удаление функций
DROP FUNCTION IF EXISTS prevent_order_archiving;