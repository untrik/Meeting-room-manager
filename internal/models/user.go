package models

import (
	"time"
)

type User struct {
	ID           string
	Email        string
	RoleID       int16
	PasswordHash *string
	CreatedAt    time.Time
}

// func (u *User) GetRoleId(role string) (int, error) {
// 	switch role {
// 	case "admin":
// 		return 1, nil
// 	case "user":
// 		return 2, nil
// 	default:
// 		return 0, fmt.Errorf("invalid role")
// 	}
// }
// func (u *User) GetRoleById(roleId int) (string, error) {
// 	switch roleId {
// 	case 1:
// 		return "admin", nil
// 	case 2:
// 		return "user", nil
// 	default:
// 		return "", fmt.Errorf("invalid role")
// 	}
// }
