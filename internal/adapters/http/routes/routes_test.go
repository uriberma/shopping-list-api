package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create router and setup routes with nil handlers for basic route testing
	router := gin.New()
	SetupRoutes(router, nil, nil)

	// Test that the router was created and routes were set up
	// We can't test individual routes with nil handlers, but we can test the setup
	assert.NotNil(t, router)

	// Test non-existent route returns 404
	req, err := http.NewRequest("GET", "/api/v1/nonexistent", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetupRoutes_HealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create router and setup routes with nil handlers for health endpoint test
	router := gin.New()
	SetupRoutes(router, nil, nil)

	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
	assert.Contains(t, w.Body.String(), "shopping-list-api")
	assert.Contains(t, w.Body.String(), "v1.0.0")
}
