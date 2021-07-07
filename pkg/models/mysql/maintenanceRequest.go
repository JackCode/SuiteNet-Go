package mysql

import (
	"database/sql"

	"github.com/jackcode/suitenet/pkg/models"
)

type MaintenanceRequestModel struct {
	DB *sql.DB
}

func (m *MaintenanceRequestModel) Insert(title, description, status string) (int, error) {
	stmt := `INSERT INTO maintenanceRequests (title, description, created, status)
			 VALUES(?, ?, UTC_TIMESTAMP(), ?)`

	result, err := m.DB.Exec(stmt, title, description, status)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *MaintenanceRequestModel) Get(id int) (*models.MaintenanceRequest, error) {
	stmt := `SELECT id, title, description, created, status FROM maintenanceRequests
	         WHERE id = ?`

	mr := &models.MaintenanceRequest{}
	err := m.DB.QueryRow(stmt, id).Scan(&mr.ID, &mr.Title, &mr.Description, &mr.Created, &mr.Status)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return mr, nil
}

func (m *MaintenanceRequestModel) OpenAndPending() ([]*models.MaintenanceRequest, error) {
	stmt := `SELECT id, title, description, created, status FROM maintenanceRequests
	         WHERE status="OPEN" OR status = "PENDING" ORDER BY created DESC`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	maintenanceRequests := []*models.MaintenanceRequest{}

	for rows.Next() {
		mr := &models.MaintenanceRequest{}

		err = rows.Scan(&mr.ID, &mr.Title, &mr.Description, &mr.Created, &mr.Status)
		if err != nil {
			return nil, err
		}
		maintenanceRequests = append(maintenanceRequests, mr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return maintenanceRequests, nil
}
