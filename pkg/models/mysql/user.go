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

func (m *UserModel) Insert(fullName, username, password, positionID, managerID string, createdByID int) error {
	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO sys_user (full_name, username, hashed_password, created, sys_user_id, position_id, manager_id, is_active)
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

	stmt := `SELECT curr_user.id, curr_user.full_name, curr_user.username, curr_user.created, curr_user.is_active, curr_user.is_clocked_in,
		     	    created_by.id AS created_by_id, created_by.full_name AS created_by_name,
				    position.id AS position_id, position.title AS position_title,
					managed_by.id AS manager_id, managed_by.full_name AS manager_name
			FROM sys_user AS curr_user
			INNER JOIN sys_user AS created_by ON curr_user.sys_user_id = created_by.id
			INNER JOIN position ON curr_user.position_id = position.id
			INNER JOIN sys_user AS managed_by ON curr_user.manager_id = managed_by.id
			WHERE curr_user.id = ? AND curr_user.is_active`

	rolesStmt := `SELECT C.id, C.title
				  FROM user_has_role AS A
				  INNER JOIN sys_user AS B ON A.sys_user_id = B.id
				  INNER JOIN site_role AS C ON A.site_role_id = C.id
				  WHERE B.id = ? AND B.is_active AND C.is_active`

	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(stmt, id).Scan(&s.ID, &s.FullName, &s.Username, &s.Created, &s.ActiveUser, &s.ClockedIn,
		&s.CreatedBy.ID, &s.CreatedBy.FullName,
		&s.Position.ID, &s.Position.Title,
		&s.Manager.ID, &s.Manager.FullName)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, models.ErrNoRecord
	} else if err != nil {
		tx.Rollback()
		return nil, err
	}

	rows, err := tx.Query(rolesStmt, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer rows.Close()

	siteRoles := []*models.SiteRole{}

	for rows.Next() {
		siteRole := &models.SiteRole{}

		err := rows.Scan(&siteRole.ID, &siteRole.Title)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		siteRoles = append(siteRoles, siteRole)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	s.SiteRoles = siteRoles
	return s, nil
}

func (m *UserModel) ChangePassword(userID int, oldPassword, newPassword string) error {
	var hashedPassword []byte
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow("SELECT hashed_password FROM sys_user WHERE id = ?", userID)
	err = row.Scan(&hashedPassword)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return models.ErrInvalidCredentials
	} else if err != nil {
		tx.Rollback()
		return err
	}

	// Check whether the hashed password and plain-text password provided match.
	// If they don't, we return the ErrInvalidCredentials error.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(oldPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		tx.Rollback()
		return models.ErrInvalidCredentials
	} else if err != nil {
		tx.Rollback()
		return err
	}

	// Create a bcrypt hash of the new plain-text password.
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt := `UPDATE sys_user
			 SET hashed_password = ?
			 WHERE id = ?`

	result, err := m.DB.Exec(stmt, hashedNewPassword, userID)
	if rows, err := result.RowsAffected(); rows == 0 || err != nil {
		tx.Rollback()
		return models.ErrInvalidCredentials
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (m *UserModel) GetActiveUsers() ([]*models.SysUser, error) {
	stmt := `SELECT id, full_name, created FROM sys_user WHERE is_active AND username != "sysadmin"`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*models.SysUser{}

	for rows.Next() {
		user := &models.SysUser{}

		err = rows.Scan(&user.ID, &user.FullName, &user.Created)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m *UserModel) ClockUser(direction string, userID int) error {
	stmt := `UPDATE sys_user SET is_clocked_in = ? WHERE id = ? AND is_active`

	isClockedIn := 0
	if direction == "in" {
		isClockedIn = 1
	}

	result, err := m.DB.Exec(stmt, isClockedIn, userID)
	if rows, err := result.RowsAffected(); rows == 0 {
		if err != nil {
			return models.ErrNoRecord
		}
	}
	if err != nil {
		return nil
	}
	return err
}
