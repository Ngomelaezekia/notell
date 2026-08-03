package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"notell/config"
	"notell/models"
	"notell/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB     *gorm.DB
	Config *config.Config
}

func NewAuthHandler(
	db *gorm.DB,
	cfg *config.Config,
) *AuthHandler {

	return &AuthHandler{
		DB:     db,
		Config: cfg,
	}
}

// =====================================================
// Session Helpers
// =====================================================

func (h *AuthHandler) createSession(
	c *gin.Context,
	user models.User,
) error {

	token, err :=
		services.GenerateToken(
			user.ID,
			user.Email,
			h.Config.JWTSecret,
		)

	if err != nil {
		return err
	}

	c.SetCookie(
		"auth_token",
		token,
		int((7 * 24 * time.Hour).Seconds()),
		"/",
		"",
		h.Config.AppEnv == "production",
		true,
	)

	return nil
}

func (h *AuthHandler) clearSession(
	c *gin.Context,
) {

	c.SetCookie(
		"auth_token",
		"",
		-1,
		"/",
		"",
		h.Config.AppEnv == "production",
		true,
	)

}

// =====================================================
// Register
// =====================================================

type registerInput struct {
	Username string `json:"username" binding:"required"`

	Email string `json:"email" binding:"required,email"`

	Password string `json:"password" binding:"required,min=6"`

	Country *string `json:"country"`

	City *string `json:"city"`

	Bio *string `json:"bio"`
}

func (h *AuthHandler) Register(
	c *gin.Context,
) {

	var input registerInput

	if err := c.ShouldBindJSON(&input); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": err.Error(),
			},
		)

		return
	}

	hash, err :=
		bcrypt.GenerateFromPassword(
			[]byte(input.Password),
			bcrypt.DefaultCost,
		)

	if err != nil {

		c.JSON(
			500,
			gin.H{
				"message": "failed hashing password",
			},
		)

		return
	}

	hashString := string(hash)

	user := models.User{

		Username: input.Username,

		Email: input.Email,

		PasswordHash: &hashString,

		Country: input.Country,

		City: input.City,

		Bio: input.Bio,
	}

	if err := h.DB.Create(&user).Error; err != nil {

		c.JSON(
			http.StatusConflict,
			gin.H{
				"message": "username or email already exists",
			},
		)

		return
	}

	if err := h.createSession(c, user); err != nil {

		c.JSON(
			500,
			gin.H{
				"message": "failed creating session",
			},
		)

		return
	}

	c.JSON(
		http.StatusCreated,
		gin.H{
			"message": "registered successfully",
		},
	)

}

// =====================================================
// Login
// =====================================================

type loginInput struct {
	Email string `json:"email" binding:"required,email"`

	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(
	c *gin.Context,
) {

	var input loginInput

	if err := c.ShouldBindJSON(&input); err != nil {

		c.JSON(
			400,
			gin.H{
				"message": err.Error(),
			},
		)

		return
	}

	var user models.User

	if err := h.DB.
		Where(
			"email = ?",
			input.Email,
		).
		First(&user).Error; err != nil {

		c.JSON(
			401,
			gin.H{
				"message": "invalid credentials",
			},
		)

		return

	}

	if user.PasswordHash == nil ||
		bcrypt.CompareHashAndPassword(
			[]byte(*user.PasswordHash),
			[]byte(input.Password),
		) != nil {

		c.JSON(
			401,
			gin.H{
				"message": "invalid credentials",
			},
		)

		return

	}

	if err := h.createSession(c, user); err != nil {

		c.JSON(
			500,
			gin.H{
				"message": "failed creating session",
			},
		)

		return
	}

	c.JSON(
		200,
		gin.H{
			"message": "login successful",
		},
	)

}

// =====================================================
// Current User
// =====================================================

func (h *AuthHandler) Me(
	c *gin.Context,
) {

	userID, exists := c.Get("userId")

	if !exists {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"message": "unauthorized",
			},
		)

		return
	}

	var user models.User

	err := h.DB.
		Select(
			"id",
			"username",
			"email",
			"profile_picture",
			"bio",
			"city",
			"country",
		).
		First(
			&user,
			userID,
		).
		Error

	if err != nil {

		c.JSON(
			http.StatusNotFound,
			gin.H{
				"message": "user not found",
			},
		)

		return

	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"data": gin.H{
				"user": user,
			},
		},
	)

}

// =====================================================
// Logout
// =====================================================

func (h *AuthHandler) Logout(
	c *gin.Context,
) {

	h.clearSession(c)

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "logged out successfully",
		},
	)

}

// =====================================================
// Google OAuth Helpers
// =====================================================

func generateStateToken() string {

	b := make(
		[]byte,
		16,
	)

	_, _ = rand.Read(b)

	return hex.EncodeToString(b)

}

func (h *AuthHandler) googleOAuthConfig() *oauth2.Config {

	return &oauth2.Config{

		ClientID: h.Config.GoogleClientID,

		ClientSecret: h.Config.GoogleClientSecret,

		RedirectURL: h.Config.GoogleRedirectURL,

		Scopes: []string{

			"https://www.googleapis.com/auth/userinfo.email",

			"https://www.googleapis.com/auth/userinfo.profile",
		},

		Endpoint: google.Endpoint,
	}

}

// =====================================================
// Google Login Redirect
// =====================================================

func (h *AuthHandler) GoogleLogin(
	c *gin.Context,
) {

	state :=
		generateStateToken()

	c.SetCookie(
		"oauth_state",
		state,
		300,
		"/",
		"",
		h.Config.AppEnv == "production",
		true,
	)

	url :=
		h.googleOAuthConfig().
			AuthCodeURL(
				state,
				oauth2.AccessTypeOffline,
			)

	c.Redirect(
		http.StatusTemporaryRedirect,
		url,
	)

}

// =====================================================
// Fetch Google User
// =====================================================

func (h *AuthHandler) getGoogleUser(
	token *oauth2.Token,
) (*GoogleUser, error,
) {

	client :=
		h.googleOAuthConfig().
			Client(
				context.Background(),
				token,
			)

	resp, err :=
		client.Get(
			"https://www.googleapis.com/oauth2/v2/userinfo",
		)

	if err != nil {

		return nil, err

	}

	defer resp.Body.Close()

	var googleUser GoogleUser

	if err :=
		json.NewDecoder(
			resp.Body,
		).
			Decode(&googleUser); err != nil {
		return nil, err
	}
	return &googleUser, nil

}

// =====================================================
// Find Or Create Google Account
// =====================================================
type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (h *AuthHandler) findOrCreateGoogleUser(
	googleUser *GoogleUser,
) (*models.User, error) {
	var user models.User
	result :=
		h.DB.
			Where(
				"google_id = ? OR email = ?",
				googleUser.ID,
				googleUser.Email,
			).
			First(&user)

	// Existing user
	if result.Error == nil {
		updates :=
			map[string]interface{}{}
		if user.GoogleID == nil {
			updates["google_id"] =
				googleUser.ID
		}

		if user.ProfilePicture == nil &&
			googleUser.Picture != "" {

			updates["profile_picture"] =
				googleUser.Picture

		}
		if len(updates) > 0 {
			h.DB.
				Model(&user).
				Updates(updates)

		}
		return &user, nil

	}
	// Create new user

	if result.Error != gorm.ErrRecordNotFound {

		return nil, result.Error

	}

	baseUsername :=
		strings.ToLower(
			strings.ReplaceAll(
				googleUser.Name,
				" ",
				"_",
			),
		)

	if baseUsername == "" {

		baseUsername =
			strings.Split(
				googleUser.Email,
				"@",
			)[0]

	}

	username :=
		baseUsername

	var count int64

	h.DB.
		Model(&models.User{}).
		Where(
			"username = ?",
			username,
		).
		Count(&count)

	if count > 0 {

		username =
			fmt.Sprintf(
				"%s_%d",
				baseUsername,
				time.Now().Unix()%10000,
			)

	}

	user = models.User{

		Username: username,

		Email: googleUser.Email,

		GoogleID: &googleUser.ID,

		ProfilePicture: &googleUser.Picture,
	}

	if err :=
		h.DB.Create(&user).Error; err != nil {

		return nil, err

	}

	return &user, nil

}

// =====================================================
// Google Callback
// =====================================================

func (h *AuthHandler) GoogleCallback(
	c *gin.Context,
) {

	cookieState, err :=
		c.Cookie(
			"oauth_state",
		)

	if err != nil ||
		cookieState != c.Query("state") {

		c.Redirect(
			http.StatusTemporaryRedirect,
			h.Config.FrontendURL+
				"/login?error=invalid_state",
		)

		return

	}

	c.SetCookie(
		"oauth_state",
		"",
		-1,
		"/",
		"",
		h.Config.AppEnv == "production",
		true,
	)

	code :=
		c.Query("code")

	if code == "" {

		c.Redirect(
			http.StatusTemporaryRedirect,
			h.Config.FrontendURL+
				"/login?error=no_code",
		)

		return

	}

	oauthToken, err :=
		h.googleOAuthConfig().
			Exchange(
				context.Background(),
				code,
			)

	if err != nil {

		c.Redirect(
			http.StatusTemporaryRedirect,
			h.Config.FrontendURL+
				"/login?error=exchange_failed",
		)

		return

	}

	googleUser, err :=
		h.getGoogleUser(
			oauthToken,
		)

	if err != nil {

		c.Redirect(
			http.StatusTemporaryRedirect,
			h.Config.FrontendURL+
				"/login?error=user_fetch_failed",
		)

		return

	}

	user, err :=
		h.findOrCreateGoogleUser(
			googleUser,
		)

	if err != nil {

		c.Redirect(
			http.StatusTemporaryRedirect,
			h.Config.FrontendURL+
				"/login?error=user_create_failed",
		)

		return

	}

	if err :=
		h.createSession(
			c,
			*user,
		); err != nil {

		c.Redirect(
			http.StatusTemporaryRedirect,
			h.Config.FrontendURL+
				"/login?error=session_failed",
		)

		return

	}

	// Redirect back without exposing JWT

	c.Redirect(
		http.StatusTemporaryRedirect,
		h.Config.FrontendURL,
	)

}
