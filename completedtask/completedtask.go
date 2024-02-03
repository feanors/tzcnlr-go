package completedtask

import (
	"errors"
	"fmt"
	"time"
)

type CompletedTask struct {
	TaskID                int       `json:"id"`
	CompanyName           string    `json:"companyName"`
	BranchName            string    `json:"branchName"`
	MachineName           string    `json:"machineName"`
	TaskStartDate         time.Time `json:"taskStartDate"`
	TaskStartTime         time.Time `json:"taskStartTime"`
	TaskEndDate           time.Time `json:"taskEndDate"`
	TaskEndTime           time.Time `json:"taskEndTime"`
	TaskDurationInMinutes int       `json:"taskDurationInMinutes"`
	IsRental              bool      `json:"isRental"`
	TaskDetail            string    `json:"taskDetail"`
}

func (ct *CompletedTask) String() string {
	return fmt.Sprintf("CompanyName: %s, BranchName: %s, TaskName: %s, TaskStartDate: %v, TaskStartTime: %v, TaskEndDate: %v, TaskEndTime: %v, TaskDurationInMinutes: %d, TaskDetail: %s, IsRental: %v\n",
		ct.CompanyName, ct.BranchName, ct.MachineName, ct.TaskStartDate, ct.TaskStartTime, ct.TaskEndDate, ct.TaskEndTime, ct.TaskDurationInMinutes, ct.TaskDetail, ct.IsRental)
}

func (ct *CompletedTask) FillDerivedCompletedTaskData() {
	if ct.TaskEndDate.IsZero() {
		ct.TaskEndDate = time.Date(ct.TaskStartDate.Year(), ct.TaskStartDate.Month(), ct.TaskStartDate.Day(), ct.TaskStartTime.Hour(), ct.TaskStartTime.Minute(), ct.TaskStartTime.Second(), ct.TaskStartTime.Nanosecond(), ct.TaskStartTime.Location()).Add(time.Minute * time.Duration(ct.TaskDurationInMinutes))
	}
	if ct.TaskEndTime.IsZero() {
		ct.TaskEndTime = ct.TaskStartTime.Add(time.Minute * time.Duration(ct.TaskDurationInMinutes))
	}
	if ct.TaskDetail == "" {
		ct.TaskDetail = "-"
	}
}

type CompletedTaskService struct {
	ctDB *CompletedTaskDB
}

func NewCompletedTaskService(ctDB *CompletedTaskDB) *CompletedTaskService {
	return &CompletedTaskService{
		ctDB: ctDB,
	}
}

func (s *CompletedTaskService) PutCompletedTask(ct CompletedTask) error {
	ct.FillDerivedCompletedTaskData()
	return s.ctDB.PutCompletedTask(ct)
}

func (s *CompletedTaskService) ValidateCompletedTaskData(ct CompletedTask) error {
	if !ct.TaskEndDate.IsZero() && ct.TaskStartDate.After(ct.TaskEndDate) {
		return errors.New("task start date before task end date")
	}
	if !ct.TaskEndTime.IsZero() && !ct.TaskEndTime.Equal(ct.TaskStartTime.Add(time.Minute*time.Duration(ct.TaskDurationInMinutes))) {
		return errors.New("task end time does not match task start time + duration in minutes")
	}
	return nil
}

func (s *CompletedTaskService) GetCompletedTasks(companyName, branchName string, startDate, endDate time.Time) ([]CompletedTask, error) {
	return s.ctDB.GetCompletedTasks(companyName, branchName, startDate, endDate)
}
