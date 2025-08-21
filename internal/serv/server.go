package serv

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"task1/internal/cache"
	"task1/internal/order"
)

type Server struct {
	cache      cache.OrderCache
	logger     *slog.Logger
	repo       order.Repository
	httpServer *http.Server
	mux        *http.ServeMux
}

func NewServer(cache cache.OrderCache, logger *slog.Logger, repo order.Repository, port int) *Server {
	mux := http.NewServeMux()
	server := &Server{
		cache:  cache,
		logger: logger,
		repo:   repo,
		mux:    mux,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}
	return server
}
func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	order, existInCache := s.cache.Load(id)
	w.Header().Set("Content-type", "application/json")
	if existInCache {
		json.NewEncoder(w).Encode(order)
		return
	}
	order, err := s.repo.FindById(context.Background(), id)
	if err != nil {
		s.logger.Error("Нету заказа в бд", "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(order)

}
func (s *Server) Start() error {
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	s.mux.HandleFunc("/getOrder", s.getOrder)
	err := s.httpServer.ListenAndServe()
	if err != nil {
		s.logger.Error("Ошибка запуска сервера", "error", err)
		return err
	}
	return nil
}
func (s *Server) Stop(ctx context.Context) {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Info("Ошибка завершения сервера", "error", err)
	}
	s.logger.Info("Сервер завершил работу")
}
