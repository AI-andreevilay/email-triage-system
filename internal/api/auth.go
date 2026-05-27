package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bzelijah/email-triage-system/internal/auth"
	"github.com/bzelijah/email-triage-system/internal/storage"
	"github.com/bzelijah/email-triage-system/internal/storage/models"
)

type telegramAuthRequest struct {
	auth.TelegramLogin
}

type telegramAuthResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	User        authUserView `json:"user"`
}

type authUserView struct {
	ID               string  `json:"id"`
	TelegramID       int64   `json:"telegram_id"`
	TelegramUsername *string `json:"telegram_username,omitempty"`
	DisplayName      *string `json:"display_name,omitempty"`
	Role             string  `json:"role"`
}

func (h *Handler) telegramAuth(w http.ResponseWriter, r *http.Request) {
	var req telegramAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := auth.VerifyTelegramLogin(req.TelegramLogin, h.authConfig.TelegramBotToken); err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid telegram login")
		return
	}

	user, err := h.store.GetUserByTelegramID(r.Context(), req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			writeJSONError(w, http.StatusForbidden, "user is not allowed")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	if !user.Enabled {
		writeJSONError(w, http.StatusForbidden, "user is disabled")
		return
	}

	token, err := h.tokens.Issue(auth.Principal{
		UserID:   user.ID,
		Role:     user.Role,
		Provider: "telegram",
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	writeJSON(w, http.StatusOK, telegramAuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.authConfig.TokenTTL.Seconds()),
		User:        userView(user),
	})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "missing principal")
		return
	}

	user, err := h.store.GetUserByID(r.Context(), principal.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			writeJSONError(w, http.StatusUnauthorized, "user not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	if !user.Enabled {
		writeJSONError(w, http.StatusForbidden, "user is disabled")
		return
	}

	writeJSON(w, http.StatusOK, userView(user))
}

func userView(user models.User) authUserView {
	return authUserView{
		ID:               user.ID,
		TelegramID:       user.TelegramID,
		TelegramUsername: user.TelegramUsername,
		DisplayName:      user.DisplayName,
		Role:             user.Role,
	}
}
