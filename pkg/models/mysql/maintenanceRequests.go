package mysql

import (
	"database/sql"

	"github.com/jackcode/suitenet/pkg/models"
)

type WorkOrderModel struct {
	DB *sql.DB
}

func (m *WorkOrderModel) Insert(title, description, status string) (int, error) {
	stmt := `INSERT INTO workOrders (title, description, created, status)
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

func (m *WorkOrderModel) Get(id int) (*models.WorkOrder, error) {
	stmt := `SELECT id, title, description, created, status FROM workOrders
	         WHERE id = ?`

	mr := &models.WorkOrder{}
	err := m.DB.QueryRow(stmt, id).Scan(&mr.ID, &mr.Title, &mr.Description, &mr.Created, &mr.Status)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return mr, nil
}

func (m *WorkOrderModel) OpenPendingInProgress() ([]*models.WorkOrder, error) {
	stmt := `SELECT id, title, description, created, status FROM workOrders
	         WHERE status="OPEN" OR status = "IN PROGRESS" OR status = "PENDING" ORDER BY created DESC`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	workOrders := []*models.WorkOrder{}

	for rows.Next() {
		mr := &models.WorkOrder{}

		err = rows.Scan(&mr.ID, &mr.Title, &mr.Description, &mr.Created, &mr.Status)
		if err != nil {
			return nil, err
		}
		workOrders = append(workOrders, mr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return workOrders, nil
}
