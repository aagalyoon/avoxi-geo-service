package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/geo-service/internal/geoip"
	"github.com/geo-service/pkg/proto"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Handler handles API requests
type Handler struct {
	geoService *geoip.Service
	logger     *logrus.Logger
}

// NewHandler creates a new API handler
func NewHandler(geoService *geoip.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		geoService: geoService,
		logger:     logger,
	}
}

// ValidateRequest represents the validation request payload
type ValidateRequest struct {
	IP               string   `json:"ip"`
	AllowedCountries []string `json:"allowed_countries"`
}

// ValidateResponse represents the validation response
type ValidateResponse struct {
	Allowed bool   `json:"allowed"`
	Country string `json:"country"`
	IP      string `json:"ip"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// RegisterHTTPRoutes registers HTTP routes
func (h *Handler) RegisterHTTPRoutes(r *mux.Router) {
	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/validate", h.handleValidate).Methods("POST")
	api.HandleFunc("/health", h.handleHealth).Methods("GET")
	
	// Serve web UI
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))
}

// RegisterGRPCServer registers the gRPC server
func (h *Handler) RegisterGRPCServer(server *grpc.Server) {
	proto.RegisterGeoServiceServer(server, h)
}

// handleValidate handles IP validation requests
func (h *Handler) handleValidate(w http.ResponseWriter, r *http.Request) {
	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.IP == "" {
		h.sendError(w, http.StatusBadRequest, "IP address is required")
		return
	}

	if len(req.AllowedCountries) == 0 {
		h.sendError(w, http.StatusBadRequest, "At least one allowed country is required")
		return
	}

	allowed, country, err := h.geoService.ValidateIP(req.IP, req.AllowedCountries)
	if err != nil {
		h.logger.WithError(err).Error("Failed to validate IP")
		h.sendError(w, http.StatusInternalServerError, "Failed to validate IP")
		return
	}

	resp := ValidateResponse{
		Allowed: allowed,
		Country: country,
		IP:      req.IP,
	}

	h.sendJSON(w, http.StatusOK, resp)
}

// handleHealth handles health check requests
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	h.sendJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// ValidateIP implements the gRPC ValidateIP method
func (h *Handler) ValidateIP(ctx context.Context, req *proto.ValidateRequest) (*proto.ValidateResponse, error) {
	allowed, country, err := h.geoService.ValidateIP(req.Ip, req.AllowedCountries)
	if err != nil {
		return nil, err
	}

	return &proto.ValidateResponse{
		Allowed: allowed,
		Country: country,
		Ip:      req.Ip,
	}, nil
}

// sendJSON sends a JSON response
func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode response")
	}
}

// sendError sends an error response
func (h *Handler) sendError(w http.ResponseWriter, status int, message string) {
	h.sendJSON(w, status, ErrorResponse{Error: message})
}