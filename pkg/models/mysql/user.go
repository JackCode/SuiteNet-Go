package mysql

import (
	"database/sql"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jackcode/suitenet/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(fullName, username, password, positionID, managerID, createdByID string) error {
	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO sys_user (full_name, username, hashed_password, created, sys_user_id, position_id, manager_id, is_active_user)
             VALUES(?, ?, ?, UTC_TIMESTAMP(), ?, ?, ?, TRUE)`

	_, err = m.DB.Exec(stmt, fullName, username, string(hashedPassword), createdByID, positionID, managerID)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "sys_user_uc_username") {
				return models.ErrDuplicateUsername
			}
		}
	}
	return err
}

func (m *UserModel) Authenticate(username, password string) (int, error) {
	// Retrieve the id and hashed password associated with the given username. If no
	// matching username exists, we return the ErrInvalidCredentials error.
	var id int
	var hashedPassword []byte
	row := m.DB.QueryRow("SELECT id, hashed_password FROM sys_user WHERE username = ?", username)
	err := row.Scan(&id, &hashedPassword)
	if err == sql.ErrNoRows {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	// Check whether the hashed password and plain-text password provided match.
	// If they don't, we return the ErrInvalidCredentials error.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	// Otherwise, the password is correct. Return the user ID.
	return id, nil
}

func (m *UserModel) Get(id int) (*models.SysUser, error) {
	s := &models.SysUser{
		CreatedBy: &models.SysUser{},
		Position:  &models.Position{},
		Manager:   &models.SysUser{},
	}

	stmt := `SELECT curr_user.id, curr_user.full_name, curr_user.username, curr_user.created, curr_user.is_active_user,
		     	    created_by.id AS created_by_id, created_by.full_name AS created_by_name,
				    position.id AS position_id, position.title AS position_title,
					managed_by.id AS manager_id, managed_by.full_name AS manager_name
			FROM sys_user AS curr_user
			INNER JOIN sys_user AS created_by ON curr_user.sys_user_id = created_by.id
			INNER JOIN position ON curr_user.position_id = position.id
			INNER JOIN sys_user AS managed_by ON curr_user.manager_id = managed_by.id
			WHERE curr_user.id = ?`

	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.FullName, &s.Username, &s.Created, &s.ActiveUser,
		&s.CreatedBy.ID, &s.CreatedBy.FullName,
		&s.Position.ID, &s.Position.Title,
		&s.Manager.ID, &s.Manager.FullName)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return s, nil
}
