package machine

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MachineDB struct {
	db *pgxpool.Pool
}

func NewMachineDB(db *pgxpool.Pool) *MachineDB {
	return &MachineDB{
		db: db,
	}
}

func (c *MachineDB) PutMachine(machineName string) error {
	query := "insert into machine (machine_name) values ($1)"
	_, err := c.db.Exec(context.Background(), query, machineName)
	return err
}

func (c *MachineDB) DeleteMachineByName(machineName string) error {
	sql := `DELETE FROM machine WHERE machine_name=$1`

	res, err := c.db.Exec(context.Background(), sql, machineName)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("machineName does not exist")
	}
	return nil
}

func (c *MachineDB) UpdateMachineByName(machineName, newMachineName string) error {
	sql := `UPDATE machine SET machine_name=$1 WHERE machine_name=$2`

	res, err := c.db.Exec(context.Background(), sql, newMachineName, machineName)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("machineName does not exist")
	}
	return nil
}

func (c *MachineDB) GetMachines() ([]Machine, error) {

	query := "select machine_id, machine_name from machine"
	var companies []Machine
	rows, err := c.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var machine Machine
		err := rows.Scan(&machine.MachineID, &machine.MachineName)

		if err != nil {
			return nil, err
		}
		companies = append(companies, machine)
	}
	return companies, nil
}
