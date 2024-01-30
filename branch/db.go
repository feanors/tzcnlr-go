package branch

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BranchDB struct {
	db *pgxpool.Pool
}

func NewBranchDB(db *pgxpool.Pool) *BranchDB {
	return &BranchDB{
		db: db,
	}
}

func (c *BranchDB) PutBranch(companyName, branchName string) error {
	query := `
        INSERT INTO branch (branch_name, company_id)
        VALUES ($1, (SELECT company_id FROM company WHERE company_name = $2))
    `
	_, err := c.db.Exec(context.Background(), query, branchName, companyName)
	return err
}

func (c *BranchDB) DeleteBranchByName(companyName, branchName string) error {
	sql := `
        DELETE FROM branch
        WHERE branch_name = $1 AND company_id = (
            SELECT company_id
            FROM company
            WHERE company_name = $2
        )
    `

	res, err := c.db.Exec(context.Background(), sql, branchName, companyName)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("no branch found with the specified name for the given company")
	}
	return nil
}

func (c *BranchDB) UpdateBranchByName(companyName, branchName, newBranchName string) error {
	sql := `
		UPDATE branch SET branch_name=$1 WHERE branch_name=$2 AND company_id = (
		    SELECT company_id
            FROM company
            WHERE company_name = $3
		)
	`

	res, err := c.db.Exec(context.Background(), sql, newBranchName, branchName, companyName)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("branchName does not exist")
	}
	return nil
}

func (c *BranchDB) GetBranches() ([]Branch, error) {
	query := "SELECT b.branch_id, b.branch_name, c.company_name FROM branch b JOIN company c ON b.company_id = c.company_id"

	var branches []Branch
	rows, err := c.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var branch Branch
		err := rows.Scan(&branch.BranchID, &branch.BranchName, &branch.CompanyName)
		if err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return branches, nil
}
