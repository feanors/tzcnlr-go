package branch

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type BranchAPI struct {
	s *BranchService
}

func NewBranchAPI(s *BranchService) *BranchAPI {
	return &BranchAPI{
		s: s,
	}
}

func (api *BranchAPI) HandleUpdateBranchByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	currentName := vars["branchName"]
	companyName := vars["companyName"]

	if currentName == "" {
		http.Error(w, "current branch name not provided in URL", http.StatusBadRequest)
		return
	}
	if companyName == "" {
		http.Error(w, "company name not provided in URL", http.StatusBadRequest)
		return
	}

	branch, ok := r.Context().Value("branch").(Branch)
	if !ok {
		http.Error(w, "error during json decode", http.StatusInternalServerError)
		return
	}

	err := api.s.UpdateBranchByName(companyName, currentName, branch)
	if err != nil {
		http.Error(w, "Error updating branch: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *BranchAPI) HandleDeleteBranchByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchName := vars["branchName"]
	companyName := vars["companyName"]

	if branchName == "" {
		http.Error(w, "branchName not set", http.StatusBadRequest)
		return
	}
	if companyName == "" {
		http.Error(w, "company name not provided in URL", http.StatusBadRequest)
		return
	}

	err := api.s.DeleteBranchByName(companyName, branchName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *BranchAPI) HandlePostBranch(w http.ResponseWriter, r *http.Request) {

	branch, ok := r.Context().Value("branch").(Branch)
	if !ok {
		http.Error(w, "error during json decode", http.StatusInternalServerError)
		return
	}

	err := api.s.PutBranch(branch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *BranchAPI) HandleGetBranch(w http.ResponseWriter, r *http.Request) {
	result, err := api.s.GetBranches()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// dont return null
	if result == nil {
		result = []Branch{}
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

func (api *BranchAPI) DecodeBranchBodyHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := r.Context().Value("body").([]byte)
		if !ok {

			http.Error(w, "error accessing the body of the request", http.StatusInternalServerError)
			return
		}

		if len(body) == 0 {
			http.Error(w, "empty request body", http.StatusBadRequest)
			return
		}

		branch, err := decodeBranch(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if branch.BranchName == "" {
			http.Error(w, "branch name not provided in request body", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "branch", branch)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func decodeBranch(body []byte) (Branch, error) {
	var branch Branch
	err := json.Unmarshal(body, &branch)
	if err != nil {
		return branch, err
	}
	return branch, nil
}
