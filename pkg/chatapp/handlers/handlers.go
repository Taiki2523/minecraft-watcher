package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"

	"github.com/taiki2523/minecraft-watcher/pkg/chatapp/auth"
	"github.com/taiki2523/minecraft-watcher/pkg/chatapp/storage"
)

const (
	sessionName   = "chatapp"
	sessionUserID = "user_id"
	sessionState  = "oauth_state"
)

type Handler struct {
	Store          *storage.Store
	Sessions       *sessions.CookieStore
	Google         *auth.GoogleClient
	FrontendOrigin string
}

type messageRequest struct {
	ChannelID int64  `json:"channel_id"`
	Body      string `json:"body"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := auth.GenerateState()
	if err != nil {
		auth.RedirectWithError(w, err)
		return
	}

	session, _ := h.Sessions.Get(r, sessionName)
	session.Values[sessionState] = state
	if err := session.Save(r, w); err != nil {
		auth.RedirectWithError(w, err)
		return
	}

	http.Redirect(w, r, h.Google.AuthCodeURL(state), http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Sessions.Get(r, sessionName)
	state, _ := session.Values[sessionState].(string)
	if state == "" || r.URL.Query().Get("state") != state {
		auth.RedirectWithError(w, errors.New("invalid oauth state"))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		auth.RedirectWithError(w, errors.New("missing code"))
		return
	}

	token, err := h.Google.Exchange(r.Context(), code)
	if err != nil {
		auth.RedirectWithError(w, err)
		return
	}

	info, err := h.Google.FetchUserInfo(r.Context(), token)
	if err != nil {
		auth.RedirectWithError(w, err)
		return
	}

	user, err := h.Store.UpsertUser(storage.User{
		GoogleID:  info.GoogleID,
		Email:     info.Email,
		Name:      info.Name,
		AvatarURL: info.AvatarURL,
	})
	if err != nil {
		auth.RedirectWithError(w, err)
		return
	}

	session.Values[sessionUserID] = user.ID
	delete(session.Values, sessionState)
	if err := session.Save(r, w); err != nil {
		auth.RedirectWithError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Sessions.Get(r, sessionName)
	delete(session.Values, sessionUserID)
	if err := session.Save(r, w); err != nil {
		auth.RedirectWithError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) Channels(w http.ResponseWriter, r *http.Request) {
	channels, err := h.Store.ListChannels()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, channels)
}

func (h *Handler) Messages(w http.ResponseWriter, r *http.Request) {
	if _, err := h.currentUser(r); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	channelID, _ := strconv.ParseInt(r.URL.Query().Get("channel_id"), 10, 64)
	if channelID == 0 {
		channelID = 1
	}

	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	messages, err := h.Store.ListMessages(channelID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, messages)
}

func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var payload messageRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}

	payload.Body = strings.TrimSpace(payload.Body)
	if payload.Body == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message cannot be empty"})
		return
	}

	if payload.ChannelID == 0 {
		payload.ChannelID = 1
	}

	message, err := h.Store.CreateMessage(payload.ChannelID, user.ID, payload.Body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, message)
}

func (h *Handler) currentUser(r *http.Request) (storage.User, error) {
	session, _ := h.Sessions.Get(r, sessionName)
	idValue, ok := session.Values[sessionUserID]
	if !ok {
		return storage.User{}, errors.New("missing session")
	}

	var id int64
	switch value := idValue.(type) {
	case int64:
		id = value
	case int:
		id = int64(value)
	case float64:
		id = int64(value)
	default:
		return storage.User{}, errors.New("invalid session")
	}

	return h.Store.GetUser(id)
}

func (h *Handler) withCORS(next http.Handler) http.Handler {
	if h.FrontendOrigin == "" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", h.FrontendOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/auth/google/login", h.Login)
	mux.HandleFunc("/auth/google/callback", h.Callback)
	mux.HandleFunc("/auth/logout", h.Logout)
	mux.Handle("/api/me", h.withCORS(http.HandlerFunc(h.Me)))
	mux.Handle("/api/channels", h.withCORS(http.HandlerFunc(h.Channels)))
	mux.Handle("/api/messages", h.withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.Messages(w, r)
		case http.MethodPost:
			h.PostMessage(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
