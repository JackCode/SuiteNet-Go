package mysql

import (
	"database/sql"
	"strings"

	"github.com/jackcode/suitenet/pkg/models"
)

type EngineeringWorkOrderModel struct {
	DB *sql.DB
}

func (m *EngineeringWorkOrderModel) Insert(title, locationID, noteContent string, createdByID int) (int, error) {
	stmt := `INSERT INTO engineering_work_order (title, created, location_id, sys_user_id, request_status_id)
			 VALUES(?, UTC_TIMESTAMP(), ?, ?, 1)`

	noteStmt := `INSERT INTO engineering_work_order_note (eng_work_order_id, content, sys_user_id, created)
	             VALUES (?, ?, ?, UTC_TIMESTAMP())`

	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec(stmt, title, locationID, createdByID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if strings.TrimSpace(noteContent) != "" {
		result, err = tx.Exec(noteStmt, int(id), noteContent, createdByID)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *EngineeringWorkOrderModel) Get(id int) (*models.EngineeringWorkOrder, error) {
	stmt := `SELECT  engineering_work_order.id, engineering_work_order.title, engineering_work_order.created, 
			         location.id, location.title,
			         sys_user.id, sys_user.full_name,
			         request_status_id, request_status.title, request_status.closed
			FROM engineering_work_order
			INNER JOIN location ON engineering_work_order.location_id = location.id
			INNER JOIN sys_user ON engineering_work_order.sys_user_id = sys_user.id
			INNER JOIN request_status ON engineering_work_order.request_status_id = request_status.id
			WHERE engineering_work_order.id = ?`

	notesStmt := `SELECT engineering_work_order_note.id, 
						 engineering_work_order_note.content, 
						 engineering_work_order_note.created, 
						 engineering_work_order_note.sys_user_id, 
						 sys_user.full_name 
				  	FROM engineering_work_order_note 
					INNER JOIN sys_user ON engineering_work_order_note.sys_user_id = sys_user.id
			  		WHERE eng_work_order_id = ? ORDER BY created DESC`

	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}

	engineeringWorkOrder := &models.EngineeringWorkOrder{
		Location:      &models.Location{},
		CreatedBy:     &models.SysUser{},
		RequestStatus: &models.RequestStatus{},
	}
	err = tx.QueryRow(stmt, id).Scan(
		&engineeringWorkOrder.ID, &engineeringWorkOrder.Title, &engineeringWorkOrder.Created,
		&engineeringWorkOrder.Location.ID, &engineeringWorkOrder.Location.Title,
		&engineeringWorkOrder.CreatedBy.ID, &engineeringWorkOrder.CreatedBy.FullName,
		&engineeringWorkOrder.RequestStatus.ID, &engineeringWorkOrder.RequestStatus.Title, &engineeringWorkOrder.RequestStatus.IsClosed)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, models.ErrNoRecord
	} else if err != nil {
		tx.Rollback()
		return nil, err
	}

	notes, err := tx.Query(notesStmt, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer notes.Close()

	workOrderNotes := []*models.EngineeringWorkOrderNote{}

	for notes.Next() {
		workOrderNote := &models.EngineeringWorkOrderNote{
			CreatedBy: &models.SysUser{},
		}
		err = notes.Scan(&workOrderNote.ID, &workOrderNote.Content, &workOrderNote.Created,
			&workOrderNote.CreatedBy.ID, &workOrderNote.CreatedBy.FullName)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		workOrderNotes = append(workOrderNotes, workOrderNote)
	}

	if err = notes.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	engineeringWorkOrder.Notes = workOrderNotes
	return engineeringWorkOrder, nil
}

func (m *EngineeringWorkOrderModel) GetIncompleteEngineeringWorkOrders() ([]*models.EngineeringWorkOrder, error) {
	stmt := `SELECT engineering_work_order.id, engineering_work_order.title, engineering_work_order.created, 
	                location.id, location.title, 
					request_status.id, request_status.title,
					sys_user.id, sys_user.full_name
			 FROM engineering_work_order
			 INNER JOIN location ON engineering_work_order.location_id = location.id
			 INNER JOIN request_status ON engineering_work_order.request_status_id = request_status.id
			 INNER JOIN sys_user ON engineering_work_order.sys_user_id = sys_user.id
			 WHERE request_status.closed != TRUE`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workOrders := []*models.EngineeringWorkOrder{}

	for rows.Next() {
		workOrder := &models.EngineeringWorkOrder{
			Location:      &models.Location{},
			RequestStatus: &models.RequestStatus{},
			CreatedBy:     &models.SysUser{},
		}
		err = rows.Scan(&workOrder.ID, &workOrder.Title, &workOrder.Created,
			&workOrder.Location.ID, &workOrder.Location.Title,
			&workOrder.RequestStatus.ID, &workOrder.RequestStatus.Title,
			&workOrder.CreatedBy.ID, &workOrder.CreatedBy.FullName)
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

func (m *EngineeringWorkOrderModel) Close(id, userID int) (*models.EngineeringWorkOrder, error) {
	stmt := `UPDATE engineering_work_order
	         SET request_status_id = 2
			 WHERE id = ?`

	noteStmt := `INSERT INTO engineering_work_order_note (eng_work_order_id, content, sys_user_id, created)
			 VALUES (?, ?, ?, UTC_TIMESTAMP())`

	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(stmt, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(noteStmt, id, "[CLOSED]", userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return m.Get(id)
}

func (m *EngineeringWorkOrderModel) Reopen(id, userID int) (*models.EngineeringWorkOrder, error) {
	stmt := `UPDATE engineering_work_order
	         SET request_status_id = 1
			 WHERE id = ?`

	noteStmt := `INSERT INTO engineering_work_order_note (eng_work_order_id, content, sys_user_id, created)
			 VALUES (?, ?, ?, UTC_TIMESTAMP())`

	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(stmt, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(noteStmt, id, "[RE-OPENED]", userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return m.Get(id)
}

func (m *EngineeringWorkOrderModel) AddNote(content string, id, userID int) (*models.EngineeringWorkOrder, error) {
	stmt := `INSERT INTO engineering_work_order_note (eng_work_order_id, content, sys_user_id, created)
			 VALUES (?, ?, ?, UTC_TIMESTAMP())`

	_, err := m.DB.Exec(stmt, id, content, userID)
	if err != nil {
		return nil, err
	}

	return m.Get(id)
}

func (m *EngineeringWorkOrderModel) GetAllWorkOrders() ([]*models.EngineeringWorkOrder, error) {
	stmt := `SELECT engineering_work_order.id, engineering_work_order.title, engineering_work_order.created, 
	                location.id, location.title, 
					request_status.id, request_status.title,
					sys_user.id, sys_user.full_name
			 FROM engineering_work_order
			 INNER JOIN location ON engineering_work_order.location_id = location.id
			 INNER JOIN request_status ON engineering_work_order.request_status_id = request_status.id
			 INNER JOIN sys_user ON engineering_work_order.sys_user_id = sys_user.id ORDER BY created DESC`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workOrders := []*models.EngineeringWorkOrder{}

	for rows.Next() {
		workOrder := &models.EngineeringWorkOrder{
			Location:      &models.Location{},
			RequestStatus: &models.RequestStatus{},
			CreatedBy:     &models.SysUser{},
		}
		err = rows.Scan(&workOrder.ID, &workOrder.Title, &workOrder.Created,
			&workOrder.Location.ID, &workOrder.Location.Title,
			&workOrder.RequestStatus.ID, &workOrder.RequestStatus.Title,
			&workOrder.CreatedBy.ID, &workOrder.CreatedBy.FullName)
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
