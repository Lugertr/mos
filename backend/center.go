package center

import (
	"errors"
	"time"
)

//----------------------------------------------------------------------------

// клиент

type LabService struct {
	ServiceID        int     `json:"service_id" db:"service_id"`
	Name             string  `json:"name" db:"name"`
	Cost             float64 `json:"cost" db:"cost"`
	ServiceCode      string  `json:"service_code" db:"service_code"`
	ExecutionTime    int     `json:"execution_time" db:"execution_time"`
	AverageDeviation float64 `json:"average_deviation" db:"average_deviation"`
}

type LabServiceUpdate struct {
	Name             *string  `json:"name"`
	Cost             *float64 `json:"cost"`
	ServiceCode      *string  `json:"service_code"`
	ExecutionTime    *int     `json:"execution_time"`
	AverageDeviation *float64 `json:"average_deviation"`
}

func (i LabServiceUpdate) Validate() error {
	if i.Name == nil && i.Cost == nil && i.ServiceCode == nil && i.ExecutionTime == nil && i.AverageDeviation == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type Patient struct {
	PatientID            int       `json:"patient_id" db:"patient_id"`
	FullName             string    `json:"full_name" db:"full_name"`
	DateOfBirth          time.Time `json:"date_of_birth" db:"date_of_birth"`
	PassportSerialNumber string    `json:"passport_serial_number" db:"passport_serial_number"`
	Phone                string    `json:"phone" db:"phone"`
	Email                string    `json:"email" db:"email"`
	InsuranceNumber      string    `json:"insurance_number" db:"insurance_number"`
	InsuranceType        string    `json:"insurance_type" db:"insurance_type"`
	InsuranceCompany     string    `json:"insurance_company" db:"insurance_company"`
}

type PatientUpdate struct {
	FullName             *string `json:"full_name"`
	DateOfBirth          *string `json:"date_of_birth"`
	PassportSerialNumber *string `json:"passport_serial_number"`
	Phone                *string `json:"phone"`
	Email                *string `json:"email"`
	InsuranceNumber      *string `json:"insurance_number"`
	InsuranceType        *string `json:"insurance_type"`
	InsuranceCompany     *int    `json:"insurance_company"`
}

func (i PatientUpdate) Validate() error {
	if i.FullName == nil && i.DateOfBirth == nil && i.PassportSerialNumber == nil && i.Phone == nil &&
		i.Email == nil && i.InsuranceNumber == nil && i.InsuranceType == nil && i.InsuranceCompany == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type InsuranceCompany struct {
	InsuranceCompanyID int    `json:"insurance_company_id" db:"insurance_company_id"`
	Name               string `json:"name" db:"name"`
	Address            string `json:"address" db:"address"`
	INN                string `json:"inn" db:"inn"`
	BankAccount        string `json:"bank_account" db:"bank_account"`
	BIK                string `json:"bik" db:"bik"`
}

type InsuranceCompanyUpdate struct {
	Name        *string `json:"name"`
	Address     *string `json:"address"`
	INN         *string `json:"inn"`
	BankAccount *string `json:"bank_account"`
	BIK         *string `json:"bik"`
}

func (i InsuranceCompanyUpdate) Validate() error {
	if i.Name == nil && i.Address == nil && i.INN == nil && i.BankAccount == nil && i.BIK == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type Order struct {
	OrderID             int       `json:"order_id" db:"order_id"`
	CreationDate        time.Time `json:"creation_date" db:"creation_date"`
	PatientID           int       `json:"patient_id" db:"patient_id"`
	StatusOrder         string    `json:"status_order" db:"status_order"`
	ExecutionTimeInDays int       `json:"execution_time_in_days" db:"execution_time_in_days"`
}

type OrderUpdate struct {
	CreationDate        *string `json:"creation_date"`
	PatientID           *int    `json:"patient_id"`
	StatusOrder         *string `json:"status_order"`
	ExecutionTimeInDays *int    `json:"execution_time_in_days"`
}

func (i OrderUpdate) Validate() error {
	if i.CreationDate == nil && i.PatientID == nil && i.StatusOrder == nil && i.ExecutionTimeInDays == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type ProvidedService struct {
	ProvidedServiceID int       `json:"provided_service_id" db:"provided_service_id"`
	ServiceID         int       `json:"service_id" db:"service_id"`
	OrderID           int       `json:"order_id" db:"order_id"`
	ExecutionDate     time.Time `json:"execution_date" db:"execution_date"`
	Performer         string    `json:"performer" db:"performer"`
}

type ProvidedServiceUpdate struct {
	ServiceID     *int    `json:"service_id"`
	OrderID       *int    `json:"order_id"`
	ExecutionDate *string `json:"execution_date"`
	Performer     *string `json:"performer"`
}

func (i ProvidedServiceUpdate) Validate() error {
	if i.ServiceID == nil && i.OrderID == nil && i.ExecutionDate == nil && i.Performer == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type Analyzer struct {
	AnalyzerID             int       `json:"analyzer_id" db:"analyzer_id"`
	OrderID                int       `json:"order_id" db:"order_id"`
	ArrivalDateTime        time.Time `json:"arrival_date_time" db:"arrival_date_time"`
	CompletionDateTime     time.Time `json:"completion_date_time" db:"completion_date_time"`
	ExecutionTimeInSeconds int       `json:"execution_time_in_seconds" db:"execution_time_in_seconds"`
}

type AnalyzerUpdate struct {
	OrderID                *int    `json:"order_id"`
	ArrivalDateTime        *string `json:"arrival_date_time"`
	CompletionDateTime     *string `json:"completion_date_time"`
	ExecutionTimeInSeconds *int    `json:"execution_time_in_seconds"`
}

func (i AnalyzerUpdate) Validate() error {
	if i.OrderID == nil && i.ArrivalDateTime == nil && i.CompletionDateTime == nil && i.ExecutionTimeInSeconds == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type Laborant struct {
	LabTechnicianID  int       `json:"lab_technician_id" db:"lab_technician_id"`
	UserId           int       `json:"user_id" db:"user_id"`
	FullName         string    `json:"full_name" db:"full_name"`
	LastLogin        time.Time `json:"last_login" db:"last_login"`
	ServicesProvided []string  `json:"services_provided" db:"services_provided"`
}

type LaborantUpdate struct {
	UserId           *int    `json:"user_id"`
	FullName         *string `json:"full_name"`
	LastLogin        *string `json:"last_login"`
	ServicesProvided *string `json:"services_provided"`
}

func (i LaborantUpdate) Validate() error {
	if i.UserId == nil && i.FullName == nil && i.LastLogin == nil && i.ServicesProvided == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type LabTechnician struct {
	LabTechnicianID  int       `json:"lab_technician_id" db:"lab_technician_id"`
	UserId           int       `json:"user_id" db:"user_id"`
	FullName         string    `json:"full_name" db:"full_name"`
	LastLogin        time.Time `json:"last_login" db:"last_login"`
	ServicesProvided []string  `json:"services_provided" db:"services_provided"`
}

type LabTechnicianUpdate struct {
	UserId           *int    `json:"user_id"`
	FullName         *string `json:"full_name"`
	LastLogin        *string `json:"last_login"`
	ServicesProvided *string `json:"services_provided"`
}

func (i LabTechnicianUpdate) Validate() error {
	if i.UserId == nil && i.FullName == nil && i.LastLogin == nil && i.ServicesProvided == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type Accountant struct {
	AccountantID int       `json:"accountant_id" db:"accountant_id"`
	UserId       int       `json:"user_id" db:"user_id"`
	FullName     string    `json:"full_name" db:"full_name"`
	LastLogin    time.Time `json:"last_login" db:"last_login"`
	Invoices     []string  `json:"invoices" db:"invoices"`
}

type UpdateAccountantInput struct {
	UserId    *int    `json:"user_id"`
	FullName  *string `json:"full_name"`
	LastLogin *string `json:"last_login"`
	Invoices  *string `json:"invoices"`
}

func (i UpdateAccountantInput) Validate() error {
	if i.UserId == nil && i.FullName == nil && i.LastLogin == nil && i.Invoices == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type Administrator struct {
	AdministratorID int `json:"administrator_id" db:"administrator_id"`
	UserId          int `json:"user_id" db:"user_id"`
}

type AdministratorUpdate struct {
	UserId *int `json:"user_id"`
}

func (i AdministratorUpdate) Validate() error {
	if i.UserId == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type ArchivedData struct {
	DataID      int       `json:"data_id" db:"data_id"`
	TableName   string    `json:"table_name" db:"table_name"`
	RecordID    int       `json:"record_id" db:"record_id"`
	ArchiveDate time.Time `json:"archive_date" db:"archive_date"`
}

type ArchivedDataUpdate struct {
	TableName   *string `json:"table_name"`
	RecordID    *int    `json:"record_id"`
	ArchiveDate *string `json:"archive_date"`
}

func (i ArchivedDataUpdate) Validate() error {
	if i.TableName == nil && i.RecordID == nil && i.ArchiveDate == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type LoginHistory struct {
	LoginHistoryID int       `json:"login_history_id" db:"login_history_id"`
	UserID         int       `json:"user_id" db:"user_id"`
	LoginTime      time.Time `json:"login_time" db:"login_time"`
	Success        bool      `json:"success" db:"success"`
}

type LoginHistoryUpdate struct {
	UserID    *int    `json:"user_id"`
	LoginTime *string `json:"login_time"`
	Success   *bool   `json:"success"`
}

func (i LoginHistoryUpdate) Validate() error {
	if i.UserID == nil && i.LoginTime == nil && i.Success == nil {
		return errors.New("update structure has no values")
	}

	return nil
}

//

type FailedLoginAttempt struct {
	AttemptID          int           `json:"attempt_id" db:"attempt_id"`
	UserID             int           `json:"user_id" db:"user_id"`
	AttemptTime        time.Time     `json:"attempt_time" db:"attempt_time"`
	IPAddress          string        `json:"ip_address" db:"ip_address"`
	CaptchaRequired    bool          `json:"captcha_required" db:"captcha_required"`
	CaptchaText        string        `json:"captcha_text" db:"captcha_text"`
	BlockedForInterval time.Duration `json:"blocked_for_interval" db:"blocked_for_interval"`
}

type FailedLoginAttemptUpdate struct {
	UserID             *int    `json:"user_id"`
	AttemptTime        *string `json:"attempt_time"`
	IPAddress          *string `json:"ip_address"`
	CaptchaRequired    *bool   `json:"captcha_required"`
	CaptchaText        *string `json:"captcha_text"`
	BlockedForInterval *string `json:"blocked_for_interval"`
}

func (i FailedLoginAttemptUpdate) Validate() error {
	if i.UserID == nil && i.AttemptTime == nil && i.IPAddress == nil && i.CaptchaRequired == nil && i.CaptchaText == nil && i.BlockedForInterval == nil {
		return errors.New("update structure has no values")
	}

	return nil
}
