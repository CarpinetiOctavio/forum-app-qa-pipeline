package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"forum-app-qa-pipeline/internal/models"
	"forum-app-qa-pipeline/internal/services"
	"github.com/gorilla/mux"
)

// PostHandler handles post HTTP requests
type PostHandler struct {
	postService *services.PostService
}

// NewPostHandler creates a new instance
func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// CreatePost handles POST /api/posts
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	// Decode the body
	var req models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respondWithError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		respondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// For simplicity, userID comes in the header
	// In production you would use JWT or sessions
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Call the service
	post, err := h.postService.CreatePost(&req, userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, post)
}

// GetAllPosts handles GET /api/posts
func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.postService.GetAllPosts()
	if err != nil {
		log.Println("GetAllPosts error:", err)
		respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, posts)
}

// GetPostByID handles GET /api/posts/{id}
func (h *PostHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	post, err := h.postService.GetPostByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, post)
}

// EditPost handles PUT /api/posts/{id}
func (h *PostHandler) EditPost(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	// Decode the body
	var req models.EditPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respondWithError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		respondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Get userID from the header
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Call the service
	post, err := h.postService.EditPost(id, &req, userID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, post)
}

// DeletePost handles DELETE /api/posts/{id}
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	// Get userID from the header
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Call the service
	err = h.postService.DeletePost(id, userID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "post deleted"})
}

// CreateComment handles POST /api/posts/{id}/comments
func (h *PostHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	// Get postID from the URL
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	// Decode the body
	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respondWithError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		respondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Get userID from the header
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Call the service
	comment, err := h.postService.CreateComment(postID, &req, userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, comment)
}

// GetComments handles GET /api/posts/{id}/comments
func (h *PostHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	// Get postID from the URL
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	comments, err := h.postService.GetCommentsByPostID(postID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, comments)
}

// EditComment handles PUT /api/posts/{postId}/comments/{commentId}
func (h *PostHandler) EditComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid post ID")
		return
	}
	commentID, err := strconv.Atoi(vars["commentId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid comment ID")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	// Decode the body
	var req models.EditCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respondWithError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		respondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	comment, err := h.postService.EditComment(postID, commentID, &req, userID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, comment)
}

// DeleteComment handles DELETE /api/posts/{postId}/comments/{commentId}
func (h *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid post ID")
		return
	}
	commentID, err := strconv.Atoi(vars["commentId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid comment ID")
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	err = h.postService.DeleteComment(postID, commentID, userID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "comment deleted"})
}
