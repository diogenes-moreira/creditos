package handler

import (
	"net/http"

	"github.com/diogenes-moreira/creditos/backend/internal/application/dto"
	"github.com/diogenes-moreira/creditos/backend/internal/application/service"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/auth"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	clientService *service.ClientService
	userRepo      port.UserRepository
	authService   *auth.LocalAuthService
}

func NewAuthHandler(clientService *service.ClientService, userRepo port.UserRepository, authService *auth.LocalAuthService) *AuthHandler {
	return &AuthHandler{
		clientService: clientService,
		userRepo:      userRepo,
		authService:   authService,
	}
}

// Register godoc
// @Summary Register a new client
// @Description Creates a new user account and client profile with the provided personal information
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	client, user, err := h.clientService.Register(
		c.Request.Context(), req.Email, req.Password,
		req.FirstName, req.LastName, req.DNI, req.CUIT, req.DateOfBirth,
		req.Phone, req.Address, req.City, req.Province, req.Country, req.IsPEP,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	token, err := h.authService.GenerateToken(user.FirebaseUID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to generate token"})
		return
	}

	_ = client
	c.JSON(http.StatusCreated, dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(user),
	})
}

// Login godoc
// @Summary Authenticate a user
// @Description Validates credentials and returns a JWT token for API access
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userRepo.FindByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid credentials"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "account is deactivated"})
		return
	}

	token, err := h.authService.GenerateToken(user.FirebaseUID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to generate token"})
		return
	}

	user.RecordLogin()
	_ = h.userRepo.Update(c.Request.Context(), user)

	c.JSON(http.StatusOK, dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(user),
	})
}
