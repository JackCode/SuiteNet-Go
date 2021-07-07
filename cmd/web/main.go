package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackcode/suitenet/pkg/models/mysql"
)

type application struct {
	errorLog            *log.Logger
	infoLog             *log.Logger
	maintenanceRequests *mysql.MaintenanceRequestModel
	templateCache       map[string]*template.Template
}

func main() {
	// Create flag for server port
	addr := flag.String("addr", ":4000", "HTTP network address")
	// Create flag for MYSQL DSN
	dsn := flag.String("dsn", "root:cvgck@/suitenet?parseTime=true", "MySQL data source name")
	// Parse flags
	flag.Parse()

	// Create info logger level
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	// Create error logger level
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:            errorLog,
		infoLog:             infoLog,
		maintenanceRequests: &mysql.MaintenanceRequestModel{DB: db},
		templateCache:       templateCache,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
