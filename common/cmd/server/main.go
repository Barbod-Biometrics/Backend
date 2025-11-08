package main

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/Barbod-Biometrics/Backened/common/internal/infrastructure/db"
)

func main() {

	// 1. Creating tools
	logger := slog.Default()

	dbConn, err := sql.Open("postgres", "postgres-url")
	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
		return
	}

	// Dependency Injection

	// 2. Creating "Porter" (Infraustructrue)
	verficationRepo := db.NewPostgresRepo(dbConn)

	// 3. Creating the use case
	// NewService takes INTERFACE, so verificationRepo fits perfectly.
	verificationSvc := usecase.NewService(verficationRepo, logger)

	// 4. Creating the handler
	verificationHandler := handler.NewVerificationHandler(verificationSvc, logger)

	http.HandleFunc("/verification", verificationHandler.HandleRequestVerification)

	logger.Info("Server starting on :8000...")
	http.ListenAndServe(":8000", nil)

}
