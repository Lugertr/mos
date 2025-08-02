package center

import "github.com/dgrijalva/jwt-go"

//1 - patient
//2 - lab_technicians
//3 - accountant
//4 - administrator

type UserType int32

const (
	PATIENT        UserType = 1
	lAB_TECHNICIAN UserType = 2
	ACCOUNTANT     UserType = 3
	ADMINISTRATOR  UserType = 4
)

type User struct {
	Id       int    `json:"-" db:"id"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	UserType int    `json:"user_type"`
}

type UserCreate struct {
	Id           int        `json:"-" db:"id"`
	Username     string     `json:"username" binding:"required"`
	Password     string     `json:"password" binding:"required"`
	UserType     int        `json:"user_type"`
	ResponceUser *jwt.Token `json:"responce_user"`
}

type UserRet struct {
	Id       int    `json:"-" db:"id"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	UserType int    `json:"user_type"`
}
