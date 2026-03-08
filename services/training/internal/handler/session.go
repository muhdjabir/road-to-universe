package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muhdjabir/road-to-universe/services/training/internal/model"
	"github.com/muhdjabir/road-to-universe/services/training/internal/repository"
	"github.com/muhdjabir/road-to-universe/services/training/internal/service"
)

type SessionHandler struct {
	svc service.SessionService
}

func NewSessionHandler(svc service.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get session"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete session"})
		return
	}

	c.Status(http.StatusNoContent)
}
