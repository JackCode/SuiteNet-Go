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
	is_active_user BOOLEAN NOT NULL
);

-- position table: Positions that sys_users are assigned to (i.e. front desk, engineer, housekeeper, etc)
CREATE TABLE position (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    department_id INTEGER NOT NULL,
    created DATETIME NOT NULL,
    sys_user_id INTEGER NOT NULL
);

-- department table: Departments that positions belong to (i.e. management, engineering, housekeeping, etc)
CREATE TABLE department (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    manager_id INTEGER NOT NULL,
    created DATETIME NOT NULL,
    sys_user_id INTEGER NOT NULL
);

-- engineering_work_order table: Contains work orders created for engineering
CREATE TABLE engineering_work_order (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    created DATETIME NOT NULL,
    location_id INTEGER NOT NULL,
    sys_user_id INTEGER NOT NULL,
    request_status_id INTEGER NOT NULL
);

-- location table: Ccontains locatiosn in hotels for engineering work orders, housekeeping requests, and others
CREATE TABLE location (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    created DATETIME NOT NULL,
    sys_user_id INTEGER NOT NULL
);

-- request_status table: Contains possible options for request stati (ie. open, complete, etc)
CREATE TABLE request_status (
	id INTEGER	NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(50) NOT NULL,
    created DATETIME NOT NULL,
    sys_user_id	INTEGER	NOT NULL
);

CREATE TABLE engineering_work_order_change (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    eng_work_order_id INTEGER NOT NULL,
    field VARCHAR(100) NOT NULL,
    old_value TEXT NOT NULL,
    new_value TEXT NOT NULL,
    sys_user_id INTEGER NOT NULL,
    created DATETIME
);

CREATE TABLE engineering_work_order_note (
	id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    eng_work_order_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    sys_user_id INTEGER NOT NULL,
    created DATETIME NOT NULL
);

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

ALTER TABLE engineering_work_order ADD CONSTRAINT workorder_in_location_fk FOREIGN KEY (location_id) REFERENCES location (id);
ALTER TABLE engineering_work_order ADD CONSTRAINT workorder_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
ALTER TABLE engineering_work_order ADD CONSTRAINT workorder_is_requeststatus_fk FOREIGN KEY (request_status_id) REFERENCES request_status (id);

ALTER TABLE location ADD CONSTRAINT location_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
ALTER TABLE location ADD CONSTRAINT location_uc_title UNIQUE (title);

ALTER TABLE request_status ADD CONSTRAINT requeststatus_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
ALTER TABLE request_status ADD CONSTRAINT requeststatus_uc_title UNIQUE (title);

ALTER TABLE engineering_work_order_change ADD CONSTRAINT change_belongsto_workorder_fk FOREIGN KEY (eng_work_order_id) REFERENCES engineering_work_order (id);
ALTER TABLE engineering_work_order_change ADD CONSTRAINT change_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);

ALTER TABLE engineering_work_order_note ADD CONSTRAINT note_belongsto_workorder_fk FOREIGN KEY (eng_work_order_id) REFERENCES engineering_work_order (id);
ALTER TABLE engineering_work_order_note ADD CONSTRAINT note_createdby_sysuser_fk FOREIGN KEY (sys_user_id) REFERENCES sys_user (id);
CREATE INDEX engworkorder_created_idx ON engineering_work_order_note(created);