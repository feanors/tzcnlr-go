package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"tzcnlr/auth"
	"tzcnlr/branch"
	"tzcnlr/company"
	"tzcnlr/completedtask"
	"tzcnlr/machine"
)

func generateSecretKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(key), nil
}

func executeMigrationSchema(filePath string, conn *pgxpool.Pool) error {
	completeQuery, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading migration file: %w", err)
	}

	queries := strings.Split(string(completeQuery), ";")

	ctx := context.Background()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	defer func() {
		// Attempt to roll back the transaction if there's an error
		if err != nil {
			if rerr := tx.Rollback(ctx); rerr != nil {
				fmt.Printf("error rolling back transaction: %v\n", rerr)
			}
		}
	}()

	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		if _, err = tx.Exec(ctx, query); err != nil {
			return fmt.Errorf("error executing query '%s': %w", query, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func DrainAndCloseRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading body: "+err.Error(), http.StatusBadRequest)
			return
		}

		err = r.Body.Close()
		if err != nil {
			http.Error(w, "error closing body: "+err.Error(), http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "body", body)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.statusCode >= 400 {
		lrw.body.Write(b)
	}
	return lrw.ResponseWriter.Write(b)
}

func ErrorLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := r.Context().Value("body").([]byte)
		if !ok {
			body = []byte{}
		}
		lrw := &LoggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lrw, r)

		if lrw.statusCode != 401 && lrw.statusCode >= 400 {
			log.Printf("HTTP Error %d: %s || Header: %s || Request Body :%s\n", lrw.statusCode, lrw.body.String(), r.URL, string(body))
		}
	})
}

func main() {

	jwtSecretKey, err := generateSecretKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println(jwtSecretKey)

	jwtKey := os.Getenv("JWT_KEY")
	databaseUrl := os.Getenv("DATABASE_URL")
	frontendURL := os.Getenv("FRONTEND_URL")
	password := os.Getenv("PASSWORD")
	port := os.Getenv("PORT_NUMBER")
	host := os.Getenv("HOST_ADDR")

	fmt.Println("CORS FRONTEND--", frontendURL, "CORS FRONTEND")
	fmt.Println("PORT--", port, "--PORT")
	fmt.Println("HOST--", host, "--HOST")

	conn, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	err = executeMigrationSchema("./mig/000000.sql", conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	completedTaskDB := completedtask.NewCompletedTaskDB(conn)
	completedTaskService := completedtask.NewCompletedTaskService(completedTaskDB)
	completedTaskApi := completedtask.NewCompletedTaskAPI(completedTaskService)

	companyDB := company.NewCompanyDB(conn)
	companyService := company.NewCompanyService(companyDB)
	companyAPI := company.NewCompanyAPI(companyService)

	machineDB := machine.NewMachineDB(conn)
	machineService := machine.NewMachineService(machineDB)
	machineAPI := machine.NewMachineAPI(machineService)

	branchDB := branch.NewBranchDB(conn)
	branchService := branch.NewBranchService(branchDB)
	branchAPI := branch.NewBranchAPI(branchService)

	authAPI := auth.NewAuthAPI("admin", password, []byte(jwtKey))

	r := mux.NewRouter()
	r.Use(DrainAndCloseRequestBody)

	corsOptions := handlers.CORS(
		handlers.AllowedOrigins([]string{frontendURL}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With", "Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"}),
	)

	r.Handle("/login", authAPI.DecodeCredentialsBodyHandler(http.HandlerFunc(authAPI.LoginHandler))).Methods("POST")

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(ErrorLoggingMiddleware)
	apiRouter.Use(authAPI.ValidateTokenMiddleware)

	apiRouter.HandleFunc("/completedTasks", completedTaskApi.HandlePostCompletedTask).Methods(http.MethodPost)
	apiRouter.HandleFunc("/completedTasks", completedTaskApi.HandleGetCompletedTask).Methods(http.MethodGet)

	apiRouter.Handle("/companies", companyAPI.DecodeCompanyBodyHandler(http.HandlerFunc(companyAPI.HandlePostCompany))).Methods(http.MethodPost)
	apiRouter.Handle("/companies/{companyName}", companyAPI.DecodeCompanyBodyHandler(http.HandlerFunc(companyAPI.HandleUpdateCompanyByName))).Methods(http.MethodPut)
	apiRouter.HandleFunc("/companies", companyAPI.HandleGetCompanies).Methods(http.MethodGet)
	apiRouter.HandleFunc("/companies/{companyName}", companyAPI.HandleDeleteCompanyByName).Methods(http.MethodDelete)

	apiRouter.Handle("/machines", machineAPI.DecodeMachineBodyHandler(http.HandlerFunc(machineAPI.HandlePostMachine))).Methods(http.MethodPost)
	apiRouter.Handle("/machines/{machineName}", machineAPI.DecodeMachineBodyHandler(http.HandlerFunc(machineAPI.HandleUpdateMachineByName))).Methods(http.MethodPut)
	apiRouter.HandleFunc("/machines", machineAPI.HandleGetMachines).Methods(http.MethodGet)
	apiRouter.HandleFunc("/machines/{machineName}", machineAPI.HandleDeleteMachineByName).Methods(http.MethodDelete)

	apiRouter.Handle("/branches", branchAPI.DecodeBranchBodyHandler(http.HandlerFunc(branchAPI.HandlePostBranch))).Methods(http.MethodPost)
	apiRouter.Handle("/branches/{companyName}/{branchName}", branchAPI.DecodeBranchBodyHandler(http.HandlerFunc(branchAPI.HandleUpdateBranchByName))).Methods(http.MethodPut)
	apiRouter.HandleFunc("/branches", branchAPI.HandleGetBranch).Methods(http.MethodGet)
	apiRouter.HandleFunc("/branches/{companyName}/{branchName}", branchAPI.HandleDeleteBranchByName).Methods(http.MethodDelete)

	err = http.ListenAndServe(host+":"+port, corsOptions(r))
	if err != nil {
		return
	}
}
