-- sys_user table: Users of the SuiteNet application
CREATE TABLE sys_user (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    full_name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    hashed_password CHAR(60) NOT NULL,
    created DATETIME NOT NULL,
    sys_user_id INTEGER NOT NULL,
    position_id INTEGER NOT NULL,
    manager_id INTEGER NOT NULL,
	is_active BOOLEAN NOT NULL
);

-- position table: Positions that sys_users are assigned to (i.e. front desk, engineer, housekeeper, etc)
CREATE TABLE position (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    department_id INTEGER NOT NULL,
    created DATETIME NOT NULL,
    sys_user_id INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL
);

-- department table: Departments that positions belong to (i.e. management, engineering, housekeeping, etc)
CREATE TABLE department (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    manager_id INTEGER NOT NULL,
    created DATETIME NOT NULL,
    sys_user_id INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL
);

-- request table: Contains requests for engineering and housekeeping
CREATE TABLE request (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    created DATETIME NOT NULL,
    location_id INTEGER NOT NULL,
    sys_user_id INTEGER NOT NULL,
    request_status_id INTEGER NOT NULL,
    request_type_id INTEGER NOT NULL
);

-- request type table: contains types of requests
CREATE TABLE request_type (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    created DATETIME NOT NULL,
    sys_user_id INTEGER NOT NULL,
    department_id INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL
);

-- location table: Ccontains locatiosn in hotels for engineering work orders, housekeeping requests, and others
CREATE TABLE location (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    created DATETIME NOT NULL,
    sys_user_id INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL
);

-- request_status table: Contains possible options for request stati (ie. open, complete, etc)
CREATE TABLE request_status (
	id INTEGER	NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(50) NOT NULL,
    created DATETIME NOT NULL,
    sys_user_id	INTEGER	NOT NULL,
    is_closed INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL
);

CREATE TABLE request_change (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    request_id INTEGER NOT NULL,
    field VARCHAR(100) NOT NULL,
    old_value TEXT NOT NULL,
    new_value TEXT NOT NULL,
    sys_user_id INTEGER NOT NULL,
    created DATETIME
);

CREATE TABLE request_note (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    request_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    sys_user_id INTEGER NOT NULL,
    created DATETIME NOT NULL
);

-- Create system admin user (default password: suitenetadmin - CHANGE THIS AFTER CREATION WITH SITE)
INSERT INTO sys_user (id, full_name, username, hashed_password, created, sys_user_id, position_id, manager_id, is_active_user)
VALUES (1, 'System Administrator', 'sysadmin', '$2y$12$/.e502VwHvjG3W/7rPpLNeB.n5.jwlJxB2v8tPvAgaq2rGiyM.DcO', UTC_TIMESTAMP(), 1, 1, 1, 1); 

ALTER TABLE sys_user ADD CONSTRAINT sysuser_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
ALTER TABLE sys_user ADD CONSTRAINT sysuser_is_position_fk FOREIGN KEY (position_id) REFERENCES position (id);
ALTER TABLE sys_user ADD CONSTRAINT sysuser_managagedby_sysuser_fk FOREIGN KEY (manager_id) REFERENCES sys_user (id);
ALTER TABLE sys_user ADD CONSTRAINT sys_user_uc_username UNIQUE (username);

ALTER TABLE position ADD CONSTRAINT position_belongsto_department_fk FOREIGN KEY (department_id) REFERENCES department (id);
ALTER TABLE position ADD CONSTRAINT position_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
ALTER TABLE position ADD CONSTRAINT position_uc_title UNIQUE (title);

ALTER TABLE department ADD CONSTRAINT department_managedby_manager_fk FOREIGN KEY (manager_id) REFERENCES sys_user (id);
ALTER TABLE department ADD CONSTRAINT department_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
ALTER TABLE department ADD CONSTRAINT department_uc_title UNIQUE (title);

ALTER TABLE request ADD CONSTRAINT request_in_location_fk FOREIGN KEY (location_id) REFERENCES location (id);
ALTER TABLE request ADD CONSTRAINT request_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
ALTER TABLE request ADD CONSTRAINT request_is_requeststatus_fk FOREIGN KEY (request_status_id) REFERENCES request_status (id);
ALTER TABLE request ADD CONSTRAINT request_is_requesttype_fk FOREIGN KEY (request_type_id) REFERENCES request_type (id);
CREATE INDEX request_created_idx ON request(created);

ALTER TABLE request_type ADD CONSTRAINT request_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
ALTER TABLE request_type ADD CONSTRAINT request_belongsto_department_fk FOREIGN KEY (department_id) REFERENCES department (id);

ALTER TABLE location ADD CONSTRAINT location_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
ALTER TABLE location ADD CONSTRAINT location_uc_title UNIQUE (title);

ALTER TABLE request_status ADD CONSTRAINT requeststatus_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
ALTER TABLE request_status ADD CONSTRAINT requeststatus_uc_title UNIQUE (title);

ALTER TABLE request_change ADD CONSTRAINT change_belongsto_request_fk FOREIGN KEY (eng_work_order_id) REFERENCES request (id);
ALTER TABLE request_change ADD CONSTRAINT change_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);

ALTER TABLE request_note ADD CONSTRAINT note_belongsto_request_fk FOREIGN KEY (eng_work_order_id) REFERENCES request (id);
ALTER TABLE request_note ADD CONSTRAINT note_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
CREATE INDEX engrequestnote_created_idx ON request_note(created);
CREATE INDEX engrequestnote_request_idx ON request_note(eng_work_order_id);

INSERT INTO department (id, title, manager_id, created, sys_user_id) 
VALUES
(1, 'Management', 1, UTC_TIMESTAMP(), 1),
(2, 'Housekeeping', 1, UTC_TIMESTAMP(), 1),
(3, 'Engineering', 1, UTC_TIMESTAMP(), 1);

INSERT INTO request_status (id, title, created, sys_user_id, is_closed, is_active)
VALUES (1, 'OPEN', UTC_TIMESTAMP(), 1, 0, 1), (2, 'CLOSED', UTC_TIMESTAMP(), 1, 1, 1);

INSERT INTO request_type (id, title, created, sys_user_id, department_id, is_active)
VALUES (1, 'Work Order', UTC_TIMESTAMP(), 1, 3, 1), (2, 'Request', UTC_TIMESTAMP(), 2, 1);