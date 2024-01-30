package company

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyDB struct {
	db *pgxpool.Pool
}

func NewCompanyDB(db *pgxpool.Pool) *CompanyDB {
	return &CompanyDB{
		db: db,
	}
}

func (c *CompanyDB) PutCompany(companyName string) error {
	query := "insert into company (company_name) values ($1)"
	_, err := c.db.Exec(context.Background(), query, companyName)
	return err
}

func (c *CompanyDB) DeleteByName(companyName string) error {
	sql := `DELETE FROM company WHERE company_name=$1`

	res, err := c.db.Exec(context.Background(), sql, companyName)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("companyName does not exist")
	}
	return nil
}

func (c *CompanyDB) UpdateCompanyByName(companyName, newCompanyName string) error {
	sql := `UPDATE company SET company_name=$1 WHERE company_name=$2`

	res, err := c.db.Exec(context.Background(), sql, newCompanyName, companyName)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("companyName does not exist")
	}
	return nil
}

func (c *CompanyDB) GetCompanies() ([]Company, error) {

	query := "select company_id, company_name from company"
	var companies []Company
	rows, err := c.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var company Company
		err := rows.Scan(&company.CompanyID, &company.CompanyName)

		if err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}
	return companies, nil
}
