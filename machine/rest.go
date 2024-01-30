package machine

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type MachineAPI struct {
	s *MachineService
}

func NewMachineAPI(s *MachineService) *MachineAPI {
	return &MachineAPI{
		s: s,
	}
}

func (api *MachineAPI) HandleUpdateMachineByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	currentName := vars["machineName"]
	if currentName == "" {
		http.Error(w, "current machine name not provided in URL", http.StatusBadRequest)
		return
	}

	machine, ok := r.Context().Value("machine").(Machine)
	if !ok {
		http.Error(w, "error during json decode", http.StatusInternalServerError)
		return
	}

	err := api.s.UpdateMachineByName(currentName, machine)
	if err != nil {
		http.Error(w, "Error updating machine: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *MachineAPI) HandleDeleteMachineByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineName := vars["machineName"]

	if machineName == "" {
		http.Error(w, "machineName not set", http.StatusBadRequest)
		return
	}

	err := api.s.DeleteMachineByName(machineName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *MachineAPI) HandlePostMachine(w http.ResponseWriter, r *http.Request) {

	machine, ok := r.Context().Value("machine").(Machine)
	if !ok {
		http.Error(w, "error during json decode", http.StatusInternalServerError)
		return
	}

	err := api.s.PutMachine(machine)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *MachineAPI) HandleGetMachines(w http.ResponseWriter, r *http.Request) {
	result, err := api.s.GetMachines()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// dont return null
	if result == nil {
		result = []Machine{}
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

func (api *MachineAPI) DecodeMachineBodyHandler(next http.Handler) http.Handler {
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

		machine, err := decodeMachine(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if machine.MachineName == "" {
			http.Error(w, "machine name not provided in request body", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "machine", machine)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func decodeMachine(body []byte) (Machine, error) {
	var machine Machine
	err := json.Unmarshal(body, &machine)
	if err != nil {
		return machine, err
	}
	return machine, nil
}
