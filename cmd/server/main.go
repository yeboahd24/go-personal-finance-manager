package main

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/yeboahd24/personal-finance-manager/internal/handler"
	"github.com/yeboahd24/personal-finance-manager/internal/middleware"
	"github.com/yeboahd24/personal-finance-manager/internal/repository"
	"github.com/yeboahd24/personal-finance-manager/internal/service"
	"github.com/yeboahd24/personal-finance-manager/internal/worker"
)

func main() {
	// Initialize DB connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	log.Println("Successfully connected to database")

	// Initialize repository
	repo := repository.NewRepository(db)

	// Initialize Plaid service
	plaidService, err := service.NewPlaidService()
	if err != nil {
		log.Fatal("Error initializing Plaid service: ", err)
	}

	// Initialize services
	userService := service.NewUserService(repo)
	accountService := service.NewAccountService(repo, plaidService)
	transactionService := service.NewTransactionService(repo, plaidService)
	categoryService := service.NewCategoryService(repo)
	budgetService := service.NewBudgetService(repo)
	analyticsService := service.NewAnalyticsService(repo)
	emailService := service.NewEmailService()
	notificationService := service.NewNotificationService(repo, emailService)
	recurringService := service.NewRecurringTransactionService(repo, transactionService)
	metricsService := service.NewMetricsService(repo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	accountHandler := handler.NewAccountHandler(accountService)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	budgetHandler := handler.NewBudgetHandler(budgetService)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	recurringHandler := handler.NewRecurringTransactionHandler(recurringService)
	metricsHandler := handler.NewSystemMetricsHandler(metricsService)

	// Initialize and start recurring transaction worker
	recurringWorker := worker.NewRecurringTransactionWorker(recurringService, notificationService, 5*time.Minute)
	go recurringWorker.Start(context.Background())

	// Create or get system user
	systemUser, err := userService.CreateUser(context.Background(), "system@personal-finance.local", "system", "System", "User")
	if err != nil {
		if strings.Contains(err.Error(), "user already exists") {
			// If user exists, get it by email
			systemUser, err = userService.GetUserByEmail(context.Background(), "system@personal-finance.local")
			if err != nil {
				log.Printf("Error getting system user: %v\n", err)
				return
			}
			log.Println("Using existing system user")
		} else {
			log.Printf("Error creating system user: %v\n", err)
			return
		}
	} else {
		log.Println("System user created successfully")
	}

	// Initialize default categories with the actual system user ID
	if err := categoryService.InitializeDefaultCategories(context.Background(), systemUser.ID); err != nil {
		log.Printf("Error initializing default categories: %v\n", err)
	} else {
		log.Println("Default categories initialized successfully")
	}

	// Process any due recurring transactions
	if err := recurringService.ProcessDueTransactions(context.Background()); err != nil {
		log.Println("Error processing recurring transactions:", err)
	}

	// Setup router
	router := setupRoutes(userHandler, accountHandler, transactionHandler, categoryHandler,
		budgetHandler, analyticsHandler, recurringHandler, metricsHandler, notificationHandler)

	// Create server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("Server error: ", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Server is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exited properly")
}

func setupRoutes(userHandler *handler.UserHandler, accountHandler *handler.AccountHandler,
	transactionHandler *handler.TransactionHandler, categoryHandler *handler.CategoryHandler,
	budgetHandler *handler.BudgetHandler, analyticsHandler *handler.AnalyticsHandler,
	recurringHandler *handler.RecurringTransactionHandler, metricsHandler *handler.SystemMetricsHandler,
	notificationHandler *handler.NotificationHandler) http.Handler {

	mux := http.NewServeMux()

	// API routes
	mux.Handle("/api/accounts", middleware.AuthMiddleware(accountHandler))
	mux.Handle("/api/accounts/", middleware.AuthMiddleware(accountHandler))
	mux.Handle("/api/transactions", middleware.AuthMiddleware(transactionHandler))
	mux.Handle("/api/transactions/", middleware.AuthMiddleware(transactionHandler))
	mux.Handle("/api/categories", middleware.AuthMiddleware(categoryHandler))
	mux.Handle("/api/categories/", middleware.AuthMiddleware(categoryHandler))
	mux.Handle("/api/budgets", middleware.AuthMiddleware(budgetHandler))
	mux.Handle("/api/budgets/", middleware.AuthMiddleware(budgetHandler))
	mux.Handle("/api/analytics", middleware.AuthMiddleware(analyticsHandler))
	mux.Handle("/api/analytics/", middleware.AuthMiddleware(analyticsHandler))

	// Auth routes
	mux.HandleFunc("/api/users", userHandler.CreateUser)
	mux.HandleFunc("/api/users/login", userHandler.Login)

	// Web routes
	templates := template.Must(template.ParseGlob("web/templates/*.html"))

	// Serve static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/js/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	}))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Public routes
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// If already authenticated, redirect to dashboard
		_, err := r.Cookie("authToken")
		if err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		data := map[string]interface{}{
			"Page": "login",
		}
		if err := templates.ExecuteTemplate(w, "layout", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		// If already authenticated, redirect to dashboard
		_, err := r.Cookie("authToken")
		if err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		data := map[string]interface{}{
			"Page": "register",
		}
		if err := templates.ExecuteTemplate(w, "layout", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Protected page routes
	protectedPages := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var page string
		switch r.URL.Path {
		case "/":
			page = "dashboard"
		case "/accounts":
			page = "accounts"
		case "/transactions":
			page = "transactions"
		case "/analytics":
			page = "analytics"
		case "/budgets":
			page = "budgets"
		default:
			http.NotFound(w, r)
			return
		}

		data := map[string]interface{}{
			"Page": page,
		}
		if err := templates.ExecuteTemplate(w, "layout", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Apply auth middleware to protected pages
	protectedHandler := middleware.AuthMiddleware(protectedPages)

	// Register protected page routes
	mux.Handle("/", protectedHandler)
	mux.Handle("/accounts", protectedHandler)
	mux.Handle("/transactions", protectedHandler)
	mux.Handle("/analytics", protectedHandler)
	mux.Handle("/budgets", protectedHandler)

	return mux
}
