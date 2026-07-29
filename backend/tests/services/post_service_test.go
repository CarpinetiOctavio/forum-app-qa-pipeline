package services

import (
	"errors"
	"testing"

	"forum-app-qa-pipeline/internal/models"
	"forum-app-qa-pipeline/internal/services"
	"forum-app-qa-pipeline/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCreatePost_Success verifies a post is created successfully
func TestCreatePost_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
	}
	mockUserRepo.On("FindByID", 1).Return(existingUser, nil)

	// Configure the mock: Create should execute successfully
	mockRepo.On("Create", mock.AnythingOfType("*models.Post")).Return(nil)

	req := &models.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post",
	}

	// ACT
	post, err := postService.CreatePost(req, 1)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "Test Post", post.Title)
	assert.Equal(t, "This is a test post", post.Content)

	// Verify the mock's methods were called
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestCreatePost_UserNotFound verifies creation fails when the given userId does not exist
func TestCreatePost_UserNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// Configure the mock: FindByID returns nil (user does not exist)
	mockUserRepo.On("FindByID", 999).Return(nil, nil)

	req := &models.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post",
	}

	// ACT
	post, err := postService.CreatePost(req, 999)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "user not found", err.Error())

	mockUserRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreatePost_RepoError verifies the error is propagated when the repository fails to create the post
func TestCreatePost_RepoError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// The user exists
	existingUser := &models.User{ID: 1, Email: "u@u.com", Username: "u"}
	mockUserRepo.On("FindByID", 1).Return(existingUser, nil)

	// The repository's Create call fails
	mockRepo.On("Create", mock.AnythingOfType("*models.Post")).Return(errors.New("db error"))

	req := &models.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post",
	}

	// ACT
	post, err := postService.CreatePost(req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "db error", err.Error())

	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestCreatePost_EmptyTitle verifies the pre-check fails when the title is empty
func TestCreatePost_EmptyTitle(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	req := &models.CreatePostRequest{
		Title:   "", // empty title
		Content: "Contenido",
	}

	// ACT
	post, err := postService.CreatePost(req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "title is required", err.Error())
	// Should NOT call the repo or the userRepo
	mockRepo.AssertNotCalled(t, "Create")
	mockUserRepo.AssertNotCalled(t, "FindByID")
}

// TestCreatePost_EmptyContent verifies the pre-check fails when the content is empty
func TestCreatePost_EmptyContent(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	req := &models.CreatePostRequest{
		Title:   "Test Post",
		Content: "", // empty content
	}

	// ACT
	post, err := postService.CreatePost(req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "content is required", err.Error())

	mockRepo.AssertNotCalled(t, "Create")
	mockUserRepo.AssertNotCalled(t, "FindByID")
}

// TestDeletePost_Success verifies the author can delete their own post
func TestDeletePost_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		Content:  "Content",
		UserID:   1, // The author is user 1
		Username: "testuser",
	}

	// Configure mocks
	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("Delete", 1).Return(nil)

	// User 1 deletes their own post
	// ACT
	err := postService.DeletePost(1, 1)

	// ASSERT
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestDeletePost_PostNotFound verifies deletion fails when the post does not exist
func TestDeletePost_PostNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// The post does not exist
	mockRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	err := postService.DeletePost(999, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "post not found", err.Error())

	// Should NOT attempt to delete
	mockRepo.AssertNotCalled(t, "Delete")
}

// TestDeletePost_NotTheAuthor verifies only the post's author can delete it
func TestDeletePost_NotTheAuthor(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		Content:  "Content",
		UserID:   1, // The author is user 1
		Username: "testuser",
	}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)

	// User 2 attempts to delete user 1's post
	// ACT
	err := postService.DeletePost(1, 2)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "you do not have permission to delete this post", err.Error())

	// Should NOT call Delete because they don't have permission
	mockRepo.AssertNotCalled(t, "Delete")
	mockRepo.AssertExpectations(t)
}

// TestDeleteComment_Success verifies the author can delete their own comment
func TestDeleteComment_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		UserID:   1,
		Username: "testuser",
	}

	existingUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
	}

	// Configure mocks
	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockUserRepo.On("FindByID", 1).Return(existingUser, nil)
	mockRepo.On("DeleteComment", 1, 10, 1).Return(nil)

	// User 1 deletes their own comment
	// ACT
	err := postService.DeleteComment(1, 10, 1)

	// ASSERT
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestDeleteComment_PostNotFound verifies comment deletion fails when the parent post does not exist
func TestDeleteComment_PostNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// The post does not exist
	mockRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	err := postService.DeleteComment(999, 10, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "post not found", err.Error())

	// Should NOT attempt to delete
	mockRepo.AssertNotCalled(t, "DeleteComment")
}

// TestDeleteComment_UserNotFound verifies comment deletion fails when the requesting user does not exist
func TestDeleteComment_UserNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		UserID:   1,
		Username: "testuser",
	}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockUserRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	err := postService.DeleteComment(1, 10, 999)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertNotCalled(t, "DeleteComment")
}

// TestDeleteComment_NotTheAuthor verifies only the comment's author can delete it
func TestDeleteComment_NotTheAuthor(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		UserID:   1,
		Username: "testuser",
	}

	existingUser := &models.User{
		ID:       2,
		Email:    "other@example.com",
		Username: "otheruser",
	}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockUserRepo.On("FindByID", 2).Return(existingUser, nil)

	// User 2 attempts to delete user 1's comment
	mockRepo.On("DeleteComment", 1, 10, 2).Return(errors.New("you do not have permission to delete this comment or it does not exist"))

	// ACT
	err := postService.DeleteComment(1, 10, 2)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "you do not have permission to delete this comment or it does not exist", err.Error())
	mockRepo.AssertExpectations(t)
}
