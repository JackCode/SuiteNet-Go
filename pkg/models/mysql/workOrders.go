package mysql

import (
	"database/sql"

	"github.com/jackcode/suitenet/pkg/models"
)

type WorkOrderModel struct {
	DB *sql.DB
}

func (m *WorkOrderModel) Insert(title, description, status, createdBy, location string) (int, error) {
	stmt := `INSERT INTO workOrders (title, description, created, status, created_by, location)
			 VALUES(?, ?, UTC_TIMESTAMP(), ?, ?, ?)`

	result, err := m.DB.Exec(stmt, title, description, status, createdBy, location)
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
	stmt := `SELECT id, title, description, created, status, created_by, location FROM workOrders
	         WHERE id = ?`

	workOrder := &models.WorkOrder{}
	err := m.DB.QueryRow(stmt, id).Scan(
		&workOrder.ID,
		&workOrder.Title,
		&workOrder.Description,
		&workOrder.Created,
		&workOrder.Status,
		&workOrder.CreatedBy,
		&workOrder.Location)

	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return workOrder, nil
}

func (m *WorkOrderModel) OpenPendingInProgress() ([]*models.WorkOrder, error) {
	stmt := `SELECT id, title, description, created, status, location FROM workOrders
	         WHERE status="OPEN" OR status = "IN PROGRESS" OR status = "PENDING" ORDER BY created DESC`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	workOrders := []*models.WorkOrder{}

	for rows.Next() {
		workOrder := &models.WorkOrder{}

		err = rows.Scan(&workOrder.ID, &workOrder.Title, &workOrder.Description, &workOrder.Created, &workOrder.Status, &workOrder.Location)
		if err != nil {
			return nil, err
		}
		workOrders = append(workOrders, workOrder)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return workOrders, nil
}

func (m *WorkOrderModel) Complete(id int) (int, error) {
	stmt := `UPDATE workorders
	         SET status = "COMPLETE"
			 WHERE id = ?`

	result, err := m.DB.Exec(stmt, id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}
