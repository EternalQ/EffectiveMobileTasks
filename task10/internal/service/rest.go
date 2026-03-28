package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/EternalQ/eff-mobile-tasks/task10/internal/models"
	"github.com/EternalQ/eff-mobile-tasks/task10/pkg/rabbit"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_duration",
		Help: "Длительность HTTP запросов",
	}, []string{"path", "method"})
)

type RestService struct {
	mu     sync.RWMutex
	users  map[int]models.User
	lastId int

	rClient *rabbit.RabbitClient
	log     *slog.Logger
}

func NewRestService(log *slog.Logger, c *rabbit.RabbitClient) *RestService {
	return &RestService{
		users:   make(map[int]models.User),
		lastId:  0,
		rClient: c,
		log:     log,
	}
}

func (s *RestService) SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	r.HandleFunc("/ws", s.handleWs)

	u := r.PathPrefix("/users").Subrouter()
	u.Use(s.metricsMw)
	u.HandleFunc("", s.RegisterUser).Methods(http.MethodPost)
	u.HandleFunc("", s.GetUsers).Methods(http.MethodGet)
	u.HandleFunc("/{id}", s.UpdateUser).Methods(http.MethodPut)
	u.HandleFunc("/{id}", s.DeleteUser).Methods(http.MethodDelete)

	return r
}

func (s *RestService) metricsMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		next.ServeHTTP(w, r)
		dur := time.Since(start)

		httpDuration.WithLabelValues(path, r.Method).Observe(float64(dur.Seconds()))

		s.log.Info("http_request",
			"path", path,
			"method", r.Method,
			"duration_ms", dur.Milliseconds(),
			"remote_ip", r.RemoteAddr,
		)
	})
}

func (s *RestService) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	s.register(&user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (s *RestService) register(user *models.User) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastId++
	user.ID = s.lastId
	s.users[user.ID] = *user

	err := s.rClient.Pub("New user: " + user.Name)
	if err != nil {
		fmt.Println("err: ", err)
	}
}

func (s *RestService) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.ID = id
	s.users[id] = user

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *RestService) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		http.Error(w, "models.User not found", http.StatusNotFound)
		return
	}

	delete(s.users, id)

	w.WriteHeader(http.StatusNoContent)
}

func (s *RestService) GetUsers(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
