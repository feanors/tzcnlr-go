package company

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
)

type CompanyAPI struct {
	s *CompanyService
}

func NewCompanyAPI(s *CompanyService) *CompanyAPI {
	return &CompanyAPI{
		s: s,
	}
}

func (api *CompanyAPI) HandleUpdateCompanyByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	currentName := vars["companyName"]
	if currentName == "" {
		err := errors.New("company name not provided in URL")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	company, ok := r.Context().Value("company").(Company)
	if !ok {
		err := errors.New("error during json decode")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := api.s.UpdateCompanyByName(currentName, company)
	if err != nil {
		http.Error(w, "Error updating company: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *CompanyAPI) HandleDeleteCompanyByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	companyName := vars["companyName"]

	if companyName == "" {
		err := errors.New("company name not set")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := api.s.DeleteCompanyByName(companyName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *CompanyAPI) HandlePostCompany(w http.ResponseWriter, r *http.Request) {

	company, ok := r.Context().Value("company").(Company)
	if !ok {
		err := errors.New("json decode error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := api.s.PutCompany(company)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *CompanyAPI) HandleGetCompanies(w http.ResponseWriter, r *http.Request) {
	result, err := api.s.GetCompanies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// dont return null
	if result == nil {
		result = []Company{}
	}

	jsonResponse, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (api *CompanyAPI) DecodeCompanyBodyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := r.Context().Value("body").([]byte)
		if !ok {
			err := errors.New("error accessing the body of the request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(body) == 0 {
			err := errors.New("empty request body")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		company, err := decodeCompany(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if company.CompanyName == "" {
			err := errors.New("company name not provided in request body")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "company", company)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func decodeCompany(body []byte) (Company, error) {
	var company Company
	err := json.Unmarshal(body, &company)
	if err != nil {
		return company, err
	}
	return company, nil
}
