package company

type Company struct {
	CompanyID   int    `json:"id"`
	CompanyName string `json:"companyName"`
}

type CompanyService struct {
	cDB *CompanyDB
}

func NewCompanyService(cDB *CompanyDB) *CompanyService {
	return &CompanyService{
		cDB: cDB,
	}
}

func (s *CompanyService) PutCompany(company Company) error {
	err := s.cDB.PutCompany(company.CompanyName)
	return err
}

func (s *CompanyService) DeleteCompanyByName(companyName string) error {
	err := s.cDB.DeleteByName(companyName)
	return err
}

func (s *CompanyService) UpdateCompanyByName(companyName string, newCompany Company) error {
	err := s.cDB.UpdateCompanyByName(companyName, newCompany.CompanyName)
	return err
}

func (s *CompanyService) GetCompanies() ([]Company, error) {
	result, err := s.cDB.GetCompanies()
	return result, err
}
