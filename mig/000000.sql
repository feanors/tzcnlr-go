CREATE TABLE IF NOT EXISTS company (
    company_id SERIAL PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    UNIQUE (company_name)
);

CREATE TABLE IF NOT EXISTS machine (
    machine_id SERIAL PRIMARY KEY,
    machine_name VARCHAR(255) NOT NULL,
    UNIQUE (machine_name)
);

CREATE TABLE IF NOT EXISTS branch (
    branch_id SERIAL PRIMARY KEY,
    branch_name VARCHAR(255) NOT NULL,
    company_id INT NOT NULL,
    FOREIGN KEY (company_id) REFERENCES company(company_id) ON DELETE CASCADE,
    UNIQUE (branch_name, company_id)
);

CREATE TABLE IF NOT EXISTS completed_task_logs (
    task_id SERIAL PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    branch_name VARCHAR(255) NOT NULL,
    machine_name VARCHAR(255) NOT NULL,
    task_start_date DATE NOT NULL,
    task_start_time TIME NOT NULL,
    task_end_date DATE NOT NULL,
    task_end_time TIME NOT NULL,
    task_duration_in_minutes INT NOT NULL,
    is_rental boolean NOT NULL,
    task_detail TEXT,
    CONSTRAINT mins_check CHECK ((task_duration_in_minutes % 30 = 0 AND is_rental != TRUE) OR (task_duration_in_minutes % 1440 = 0 AND is_rental = TRUE)),
    UNIQUE (company_name, machine_name, branch_name, task_start_date, task_start_time, task_duration_in_minutes, is_rental)
);
