package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/samAlcazar/gama_api/internal/config"
	"github.com/samAlcazar/gama_api/internal/db"
	"github.com/samAlcazar/gama_api/internal/handler"
	"github.com/samAlcazar/gama_api/internal/middleware"
	"github.com/samAlcazar/gama_api/internal/repository"
	"github.com/samAlcazar/gama_api/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	defer database.Close()
	log.Println("Conexión a la base de datos establecida correctamente")

	userRepo := repository.NewUserRepository(database)
	deptRepo := repository.NewDepartmentRepository(database)

	authService := service.NewAuthService(cfg, userRepo)
	deptService := service.NewDepartmentService(deptRepo)
	userService := service.NewUserService(userRepo, deptRepo)

	authHandler := handler.NewAuthHandler(authService)
	deptHandler := handler.NewDepartmentHandler(deptService)
	userHandler := handler.NewUserHandler(userService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	authMW := middleware.AuthMiddleware(cfg)

	mux.Handle("GET /api/v1/auth/me", authMW(http.HandlerFunc(authHandler.Me)))

	mux.Handle("GET /api/v1/departments", authMW(http.HandlerFunc(deptHandler.List)))
	mux.Handle("POST /api/v1/departments", authMW(http.HandlerFunc(deptHandler.Create)))

	mux.Handle("GET /api/v1/users", authMW(
		middleware.RequirePermission("USUARIO_VER")(http.HandlerFunc(userHandler.List)),
	))
	mux.Handle("GET /api/v1/users/{id}", authMW(
		middleware.RequirePermission("USUARIO_VER")(http.HandlerFunc(userHandler.GetByID)),
	))
	mux.Handle("POST /api/v1/users", authMW(
		middleware.RequirePermission("USUARIO_CREAR")(http.HandlerFunc(userHandler.Create)),
	))
	mux.Handle("PUT /api/v1/users/{id}", authMW(
		middleware.RequirePermission("USUARIO_EDITAR")(http.HandlerFunc(userHandler.Update)),
	))
	mux.Handle("DELETE /api/v1/users/{id}", authMW(
		middleware.RequirePermission("USUARIO_DESACTIVAR")(http.HandlerFunc(userHandler.Deactivate)),
	))

	adminTestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Acceso concedido al panel de auditoría (permiso validado correctamente)"}`))
	})
	mux.Handle("GET /api/v1/admin/audit", authMW(
		middleware.RequirePermission("AUDITORIA_VER")(adminTestHandler),
	))

	serverAddr := ":" + cfg.Port
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("Servidor HTTP escuchando en el puerto %s...", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error en el servidor HTTP: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Apagando el servidor HTTP de forma segura...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error al apagar el servidor: %v", err)
	}

	log.Println("Servidor HTTP apagado correctamente")
}
