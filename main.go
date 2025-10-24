package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"taskflow/api"
	"taskflow/store"
	"taskflow/worker"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	task_store := store.NewTaskStore()
	job_dispatcher := worker.NewDispatcher(task_store, 0)
	job_dispatcher.Start()

	task_api := api.NewTaskAPI(task_store, job_dispatcher)

	r.Use(
		middleware.Logger,
		middleware.Recoverer,
	)

	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", task_api.GetTasksHandler)
		r.Get("/{id}", task_api.GetTaskHandler)
		r.Post("/", task_api.CreateTaskHandler)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	stopchan := make(chan os.Signal, 1)
	signal.Notify(stopchan, os.Interrupt)

	go func() {
		log.Println("Server is starting")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("ERROR: server error: %v", err)
		}
	}()

	<-stopchan
	log.Println("Shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("ERROR: Server forced to shutdown: %v", err)
	}

	job_dispatcher.ShutDown()
	log.Printf("Server and dispatcher shutdown greacefully")
}
