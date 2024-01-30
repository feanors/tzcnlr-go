package completedtask

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	_ "time/tzdata"
)

type CompletedTaskAPI struct {
	s *CompletedTaskService
}

func NewCompletedTaskAPI(s *CompletedTaskService) *CompletedTaskAPI {
	return &CompletedTaskAPI{
		s: s,
	}
}

func (api *CompletedTaskAPI) HandlePostCompletedTask(w http.ResponseWriter, r *http.Request) {
	body, ok := r.Context().Value("body").([]byte)
	if !ok {
		http.Error(w, "error accessing the body of the request", http.StatusInternalServerError)
		return
	}

	// should be merged with check missing bodyInput maybe idk
	if len(body) == 0 {
		http.Error(w, "empty request body", http.StatusBadRequest)
		return
	}

	ct, err := decodeCompletedTask(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = checkMissingBodyInput(ct); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = api.s.ValidateCompletedTaskData(ct); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = api.s.PutCompletedTask(ct); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (api *CompletedTaskAPI) HandleGetCompletedTask(w http.ResponseWriter, r *http.Request) {
	companyName := r.URL.Query().Get("companyName")
	branchName := r.URL.Query().Get("branchName")

	startDate, err := parseDate(r.URL.Query().Get("startDate"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	endDate, err := parseDate(r.URL.Query().Get("endDate"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := api.s.GetCompletedTasks(companyName, branchName, startDate, endDate)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// dont return null
	if result == nil {
		result = []CompletedTask{}
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

func parseDate(date string) (time.Time, error) {
	var zeroDate time.Time
	if date == "" {
		return zeroDate, nil
	}

	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, date)
	if err != nil {
		return zeroDate, err
	}

	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		return zeroDate, fmt.Errorf("date&time conversion error: %w", err)
	}

	return parsedDate.In(loc), nil
}

func decodeCompletedTask(body []byte) (CompletedTask, error) {
	var ct CompletedTask
	err := json.Unmarshal(body, &ct)
	if err != nil {
		return ct, fmt.Errorf("decode error: %w", err)
	}

	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		return ct, fmt.Errorf("date&time conversion error: %w", err)
	}

	ct.TaskStartDate = ct.TaskStartDate.In(loc)
	ct.TaskStartTime = ct.TaskStartTime.In(loc)

	return ct, nil
}

func checkMissingBodyInput(ct CompletedTask) error {
	if ct.CompanyName == "" {
		return errors.New("company name not set")
	}
	if ct.BranchName == "" {
		return errors.New("branch name not set")
	}
	if ct.MachineName == "" {
		return errors.New("machine name not set")
	}
	if ct.TaskStartDate.IsZero() {
		return errors.New("task start date not set")
	}
	if ct.TaskStartTime.IsZero() {
		return errors.New("task start time not set")
	}
	if ct.TaskDurationInMinutes == 0 {
		return errors.New("task duration not set")
	}
	return nil
}
