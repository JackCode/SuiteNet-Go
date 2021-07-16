package mysql

import (
	"database/sql"
	"strings"

	"github.com/jackcode/suitenet/pkg/models"
)

type RequestModel struct {
	DB *sql.DB
}

func (m *RequestModel) Insert(title, locationID, noteContent, department string, createdByID int) (int, error) {
	stmt := `INSERT INTO request (title, created, location_id, sys_user_id, request_status_id, department_id)
			 VALUES(?, UTC_TIMESTAMP(), ?, ?, 1, 
			 (SELECT id FROM department WHERE title = ?))`

	noteStmt := `INSERT INTO request_note (request_id, content, sys_user_id, created)
	             VALUES (?, ?, ?, UTC_TIMESTAMP())`

	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec(stmt, title, locationID, createdByID, department)
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
		result, err = tx.Exec(noteStmt, int(id), strings.TrimSpace(noteContent), createdByID)
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

func (m *RequestModel) Get(id int, department string) (*models.Request, error) {
	stmt := `SELECT  request.id, request.title, request.created, 
			         location.id, location.title,
			         sys_user.id, sys_user.full_name,
			         request_status_id, request_status.title, request_status.is_closed,
					 department.id, department.title
			FROM request
			INNER JOIN location ON request.location_id = location.id
			INNER JOIN sys_user ON request.sys_user_id = sys_user.id
			INNER JOIN request_status ON request.request_status_id = request_status.id
			INNER JOIN department ON request.department_id = department.id
			WHERE request.id = ? 
				AND request.department_id = (SELECT id FROM department WHERE title = ?)`

	notesStmt := `SELECT request_note.id, 
						 request_note.content, 
						 request_note.created, 
						 request_note.sys_user_id, 
						 sys_user.full_name 
				  	FROM request_note 
					INNER JOIN sys_user ON request_note.sys_user_id = sys_user.id
			  		WHERE request_id = ? ORDER BY created DESC`

	readStmt := `SELECT su.id, su.full_name 
				 FROM request_read
				 INNER JOIN sys_user AS su ON request_read.sys_user_id = su.id
				 WHERE request_read.request_id = ? AND su.is_active`

	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}

	request := &models.Request{
		Location:      &models.Location{},
		CreatedBy:     &models.SysUser{},
		RequestStatus: &models.RequestStatus{},
		Department:    &models.Department{},
	}

	err = tx.QueryRow(stmt, id, department).Scan(
		&request.ID, &request.Title, &request.Created,
		&request.Location.ID, &request.Location.Title,
		&request.CreatedBy.ID, &request.CreatedBy.FullName,
		&request.RequestStatus.ID, &request.RequestStatus.Title, &request.RequestStatus.IsClosed,
		&request.Department.ID, &request.Department.Title,
	)
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

	requestNotes := []*models.RequestNote{}

	for notes.Next() {
		requestNote := &models.RequestNote{
			CreatedBy: &models.SysUser{},
		}
		err = notes.Scan(&requestNote.ID, &requestNote.Content, &requestNote.Created,
			&requestNote.CreatedBy.ID, &requestNote.CreatedBy.FullName)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		requestNotes = append(requestNotes, requestNote)
	}

	if err = notes.Err(); err != nil {
		return nil, err
	}

	rows, err := tx.Query(readStmt, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer rows.Close()

	users := []*models.SysUser{}

	for rows.Next() {
		user := &models.SysUser{}
		err = rows.Scan(&user.ID, &user.FullName)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	request.Notes = requestNotes
	request.ReadBy = users
	return request, nil
}

func (m *RequestModel) GetIncompleteRequests(department string) ([]*models.Request, error) {
	stmt := `SELECT request.id, request.title, request.created, 
	                location.id, location.title, 
					request_status.id, request_status.title,
					sys_user.id, sys_user.full_name
			 FROM request
			 INNER JOIN location ON request.location_id = location.id
			 INNER JOIN request_status ON request.request_status_id = request_status.id
			 INNER JOIN sys_user ON request.sys_user_id = sys_user.id
			 WHERE request_status.is_closed != TRUE
			 AND department_id = 
		     	(SELECT department.id FROM department WHERE department.title = ?)`

	rows, err := m.DB.Query(stmt, department)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []*models.Request{}

	for rows.Next() {
		request := &models.Request{
			Location:      &models.Location{},
			RequestStatus: &models.RequestStatus{},
			CreatedBy:     &models.SysUser{},
		}
		err = rows.Scan(&request.ID, &request.Title, &request.Created,
			&request.Location.ID, &request.Location.Title,
			&request.RequestStatus.ID, &request.RequestStatus.Title,
			&request.CreatedBy.ID, &request.CreatedBy.FullName)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

func (m *RequestModel) Close(id, userID int, department string) (*models.Request, error) {
	stmt := `UPDATE request
	         SET request_status_id = 2
			 WHERE id = ?
			 AND department_id = (SELECT id FROM department WHERE title = ?)`

	noteStmt := `INSERT INTO request_note (request_id, content, sys_user_id, created)
			 VALUES (?, ?, ?, UTC_TIMESTAMP())`

	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(stmt, id, department)
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
	return m.Get(id, department)
}

func (m *RequestModel) Reopen(id, userID int, department string) (*models.Request, error) {
	stmt := `UPDATE request
	         SET request_status_id = 1
			 WHERE id = ?
			 AND department_id = (SELECT id FROM department WHERE title = ?)`

	noteStmt := `INSERT INTO request_note (request_id, content, sys_user_id, created)
			 VALUES (?, ?, ?, UTC_TIMESTAMP())`

	tx, err := m.DB.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(stmt, id, department)
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
	return m.Get(id, department)
}

func (m *RequestModel) AddNote(content, department string, id, userID int) (*models.Request, error) {
	stmt := `INSERT INTO request_note (request_id, content, sys_user_id, created)
			 VALUES (?, ?, ?, UTC_TIMESTAMP())`

	_, err := m.DB.Exec(stmt, id, content, userID)
	if err != nil {
		return nil, err
	}

	return m.Get(id, department)
}

func (m *RequestModel) GetAllRequests(department string) ([]*models.Request, error) {
	stmt := `SELECT request.id, request.title, request.created, 
	                location.id, location.title, 
					request_status.id, request_status.title,
					sys_user.id, sys_user.full_name
			 FROM request
			 INNER JOIN location ON request.location_id = location.id
			 INNER JOIN request_status ON request.request_status_id = request_status.id
			 INNER JOIN sys_user ON request.sys_user_id = sys_user.id 
			 WHERE request.department_id = 
			 		(SELECT department.id FROM department WHERE department.title = ?) 
			 ORDER BY created DESC`

	rows, err := m.DB.Query(stmt, department)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []*models.Request{}

	for rows.Next() {
		request := &models.Request{
			Location:      &models.Location{},
			RequestStatus: &models.RequestStatus{},
			CreatedBy:     &models.SysUser{},
		}
		err = rows.Scan(&request.ID, &request.Title, &request.Created,
			&request.Location.ID, &request.Location.Title,
			&request.RequestStatus.ID, &request.RequestStatus.Title,
			&request.CreatedBy.ID, &request.CreatedBy.FullName)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

func (m *RequestModel) MarkRead(requestID, userID int) error {
	stmt := `INSERT INTO request_read (request_id, sys_user_id, created)
		     SELECT ?, ?, UTC_TIMESTAMP()
	         FROM dual
	         WHERE NOT EXISTS (SELECT 1 
						       FROM request_read 
						       WHERE request_id = ? AND sys_user_id = ?)`

	_, err := m.DB.Exec(stmt, requestID, userID, requestID, userID)
	if err != nil {
		return err
	}

	return err
}
