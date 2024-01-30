package machine

type Machine struct {
	MachineID   int    `json:"id"`
	MachineName string `json:"machineName"`
}

type MachineService struct {
	cDB *MachineDB
}

func NewMachineService(cDB *MachineDB) *MachineService {
	return &MachineService{
		cDB: cDB,
	}
}

func (s *MachineService) PutMachine(machine Machine) error {
	err := s.cDB.PutMachine(machine.MachineName)
	return err
}

func (s *MachineService) DeleteMachineByName(machineName string) error {
	err := s.cDB.DeleteMachineByName(machineName)
	return err
}

func (s *MachineService) UpdateMachineByName(machineName string, newMachine Machine) error {
	err := s.cDB.UpdateMachineByName(machineName, newMachine.MachineName)
	return err
}

func (s *MachineService) GetMachines() ([]Machine, error) {
	result, err := s.cDB.GetMachines()
	return result, err
}
