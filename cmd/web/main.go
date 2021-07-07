package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackcode/suitenet/pkg/models/mysql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

type application struct {
	errorLog            *log.Logger
	infoLog             *log.Logger
	session             *sessions.Session
	maintenanceRequests *mysql.MaintenanceRequestModel
	templateCache       map[string]*template.Template
}

func main() {
	// Create flag for server port
	addr := flag.String("addr", ":4000", "HTTP network address")
	// Create flag for MYSQL DSN
	dsn := flag.String("dsn", "root:cvgck@/suitenet?parseTime=true", "MySQL data source name")
	// Define a new command-line flag for the session secret (a random key which
	// will be used to encrypt and authenticate session cookies). It should be 32
	// bytes long.
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")
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

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true

	app := &application{
		errorLog:            errorLog,
		infoLog:             infoLog,
		session:             session,
		maintenanceRequests: &mysql.MaintenanceRequestModel{DB: db},
		templateCache:       templateCache,
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
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
