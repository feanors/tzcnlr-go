package completedtask

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func NewCompletedTaskDB(db *pgxpool.Pool) *CompletedTaskDB {
	return &CompletedTaskDB{
		db: db,
	}
}

type CompletedTaskDB struct {
	db *pgxpool.Pool
}

func (c *CompletedTaskDB) PutCompletedTask(ct CompletedTask) error {
	query := `
	INSERT INTO completed_task_logs (
    	company_name, branch_name, machine_name, task_start_date, task_start_time, task_end_date, task_end_time, task_duration_in_minutes, is_rental, task_detail
	) 
	VALUES (
		(SELECT company_name FROM company WHERE company_name = $1),
		(SELECT branch_name FROM branch WHERE branch_name = $2 AND company_id = (SELECT company_id FROM company WHERE company_name = $1)),
		(SELECT machine_name FROM machine WHERE machine_name = $3),
		$4, $5, $6, $7, $8, $9, $10 
	)`

	/* Following query is probably more performant but fails big time on type deduce mismatch
		`
	        INSERT INTO completed_task_logs (
	            company_name, branch_name, machine_name, task_start_date, task_start_time,
	            task_end_date, task_end_time, task_duration_in_minutes
	        )
	        SELECT
	            $1, $2, $3,
	            $4, $5, $6, $7, $8
	        FROM
	            company c, branch b, machine m
	        WHERE
	            c.company_name = $1
	            AND b.branch_name = $2
	            AND m.machine_name = $3
	            AND b.company_id = c.company_id
	        LIMIT 1
		`
	*/

	_, err := c.db.Exec(
		context.Background(),
		query,
		ct.CompanyName,
		ct.BranchName,
		ct.MachineName,
		ct.TaskStartDate,
		ct.TaskStartTime,
		ct.TaskEndDate,
		ct.TaskEndTime,
		ct.TaskDurationInMinutes,
		ct.IsRental,
		ct.TaskDetail,
	)
	return err
}

func (c *CompletedTaskDB) GetCompletedTasks(companyName, branchName string, startDate, endDate time.Time) ([]CompletedTask, error) {

	queryData := buildFilteredQuery(companyName, branchName, startDate, endDate)
	var completedTasks []CompletedTask
	query, params := queryData.query, queryData.params

	rows, err := c.db.Query(context.Background(), query, params...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var completedTask CompletedTask
		err := rows.Scan(
			&completedTask.TaskID,
			&completedTask.CompanyName,
			&completedTask.BranchName,
			&completedTask.MachineName,
			&completedTask.TaskStartDate,
			&completedTask.TaskStartTime,
			&completedTask.TaskEndDate,
			&completedTask.TaskEndTime,
			&completedTask.TaskDurationInMinutes,
			&completedTask.IsRental,
			&completedTask.TaskDetail)

		if err != nil {
			return nil, err
		}
		completedTasks = append(completedTasks, completedTask)
	}

	return completedTasks, nil
}

type queryData struct {
	query  string
	params []interface{}
}

func buildFilteredQuery(companyName, branchName string, startDate, endDate time.Time) queryData {
	query := "SELECT " +
		"task_id, company_name, branch_name, machine_name, task_start_date, task_start_time, task_end_date, task_end_time, task_duration_in_minutes, is_rental, task_detail" +
		" FROM completed_task_logs" +
		" WHERE 1=1"
	params := []interface{}{}
	paramCount := 1

	if companyName != "" {
		query += fmt.Sprintf(" AND company_name = $%d", paramCount)
		params = append(params, companyName)
		paramCount++
	}
	if branchName != "" {
		query += fmt.Sprintf(" AND branch_name = $%d", paramCount)
		params = append(params, branchName)
		paramCount++
	}
	if !startDate.IsZero() {
		query += fmt.Sprintf(" AND task_start_date >= $%d", paramCount)
		params = append(params, startDate)
		paramCount++
	}
	if !endDate.IsZero() {
		query += fmt.Sprintf(" AND task_start_date <= $%d", paramCount)
		params = append(params, endDate)
		paramCount++
	}

	return queryData{query: query, params: params}
}
