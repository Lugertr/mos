DROP TRIGGER IF EXISTS client_deletion_trigger ON client_table;
DROP TRIGGER IF EXISTS insert_client_trigger ON client_table;
DROP TRIGGER IF EXISTS update_client_trigger ON client_table;
DROP TRIGGER IF EXISTS delete_client_trigger ON client_table;
DROP TRIGGER IF EXISTS date_check_trigger ON client_table;
DROP TRIGGER IF EXISTS service_table_check_service_type_exists ON service_table;

DROP FUNCTION IF EXISTS get_client_info(cur_client_id numeric); --Вернуть чек клиента
DROP FUNCTION IF EXISTS get_client_services(client_id numeric);
DROP FUNCTION IF EXISTS client_deletion_trigger_function();
DROP FUNCTION IF EXISTS count_days(client_id numeric);
DROP FUNCTION IF EXISTS get_service_types();    --Вернуть неиспользующиеся спросом услуги
DROP FUNCTION IF EXISTS get_today_tomorrow();   -- Вернуть номера которые освободяться сегодня или завтра
DROP FUNCTION IF EXISTS update_app_table_on_insert();
DROP FUNCTION IF EXISTS update_app_table_on_update();
DROP FUNCTION IF EXISTS update_app_table_on_delete();
DROP FUNCTION IF EXISTS check_date();
DROP FUNCTION IF EXISTS check_service_type_exists();

DROP TABLE IF EXISTS service_table;
DROP TABLE IF EXISTS service_type_table;
DROP TABLE IF EXISTS old_services_table;
DROP TABLE IF EXISTS client_table;
DROP TABLE IF EXISTS old_client_table;
DROP TABLE IF EXISTS app_table;
DROP TABLE IF EXISTS app_type_table;
DROP TABLE IF EXISTS users;