package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhdjabir/road-to-universe/services/training/internal/model"
	"github.com/muhdjabir/road-to-universe/services/training/internal/repository"
	"github.com/muhdjabir/road-to-universe/services/training/internal/service"
	"go.uber.org/zap"
)

type SessionHandler struct {
	svc service.SessionService
	log *zap.Logger
}

func NewSessionHandler(svc service.SessionService, log *zap.Logger) *SessionHandler {
	return &SessionHandler{svc: svc, log: log}
}

// POST /api/v1/training
func (h *SessionHandler) Create(c *gin.Context) {
	var req model.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO Phase 2: extract userID from JWT claims
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "00000000-0000-0000-0000-000000000001"
	}

	session, err := h.svc.CreateSession(c.Request.Context(), userID, &req)
	if err != nil {
		h.log.Error("failed to create session", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GET /api/v1/training
func (h *SessionHandler) List(c *gin.Context) {
	// TODO Phase 2: extract userID from JWT claims
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "00000000-0000-0000-0000-000000000001"
	}

	sessions, err := h.svc.ListSessions(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("failed to list sessions", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sessions"})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// GET /api/v1/training/:id
func (h *SessionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	session, err := h.svc.GetSession(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		h.log.Error("failed to get session", zap.String("session_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get session"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// PUT /api/v1/training/:id
func (h *SessionHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req model.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.svc.UpdateSession(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		h.log.Error("failed to update session", zap.String("session_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update session"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// DELETE /api/v1/training/:id
func (h *SessionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteSession(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		h.log.Error("failed to delete session", zap.String("session_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete session"})
		return
	}

	c.Status(http.StatusNoContent)
}
