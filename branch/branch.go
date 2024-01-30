package branch

type Branch struct {
	BranchID    int    `json:"id"`
	BranchName  string `json:"branchName"`
	CompanyName string `json:"companyName"`
}

type BranchService struct {
	cDB *BranchDB
}

func NewBranchService(cDB *BranchDB) *BranchService {
	return &BranchService{
		cDB: cDB,
	}
}

func (s *BranchService) PutBranch(branch Branch) error {
	err := s.cDB.PutBranch(branch.CompanyName, branch.BranchName)
	return err
}

func (s *BranchService) DeleteBranchByName(companyName, branchName string) error {
	err := s.cDB.DeleteBranchByName(companyName, branchName)
	return err
}

func (s *BranchService) UpdateBranchByName(companyName, branchName string, newBranch Branch) error {
	err := s.cDB.UpdateBranchByName(companyName, branchName, newBranch.BranchName)
	return err
}

func (s *BranchService) GetBranches() ([]Branch, error) {
	result, err := s.cDB.GetBranches()
	return result, err
}
