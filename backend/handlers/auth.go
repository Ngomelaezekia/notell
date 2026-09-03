package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
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

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{DB: db, Config: cfg}
}

const sessionLifetime = 72 * time.Hour

func (h *AuthHandler) setSessionCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", "", h.Config.AppEnv == "production", true)
}

func (h *AuthHandler) createSession(c *gin.Context, user models.User) error {
	token, err := services.GenerateToken(user.ID, user.Email, h.Config.JWTSecret)
	if err != nil {
		return err
	}
	h.setSessionCookie(c, "auth_token", token, int(sessionLifetime.Seconds()))
	return nil
}

func (h *AuthHandler) clearSession(c *gin.Context) {
	h.setSessionCookie(c, "auth_token", "", -1)
}

type registerInput struct {
	Username string  `json:"username" binding:"required,max=50"`
	Email    string  `json:"email" binding:"required,email,max=254"`
	Password string  `json:"password" binding:"required,min=6,max=72"`
	Country  *string `json:"country" binding:"max=100"`
	City     *string `json:"city" binding:"max=100"`
	Bio      *string `json:"bio" binding:"max=2000"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input registerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed hashing password"})
		return
	}
	hashString := string(hash)
	user := models.User{
		Username:     strings.TrimSpace(input.Username),
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: &hashString,
		Country:      input.Country,
		City:         input.City,
		Bio:          input.Bio,
	}
	if user.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "username cannot be empty"})
		return
	}
	if err := h.DB.Create(&user).Error; err != nil {
		if isUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"message": "username or email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create account"})
		return
	}
	if err := h.createSession(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed creating session"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "registered successfully"})
}

type loginInput struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,max=72"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input loginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", strings.ToLower(strings.TrimSpace(input.Email))).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
		return
	}
	if user.PasswordHash == nil || bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
		return
	}
	if err := h.createSession(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed creating session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	var user models.User
	err := h.DB.Select("id", "username", "email", "profile_picture", "bio", "city", "country", "status", "allow_followers", "created_at", "updated_at").First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "session user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load current user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"user": user}})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearSession(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func generateStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate oauth state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (h *AuthHandler) googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.Config.GoogleClientID,
		ClientSecret: h.Config.GoogleClientSecret,
		RedirectURL:  h.Config.GoogleRedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state, err := generateStateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to start Google authentication"})
		return
	}
	h.setSessionCookie(c, "oauth_state", state, 300)
	c.Redirect(http.StatusTemporaryRedirect, h.googleOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline))
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func (h *AuthHandler) getGoogleUser(ctx context.Context, token *oauth2.Token) (*GoogleUser, error) {
	client := h.googleOAuthConfig().Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, err
	}
	if googleUser.ID == "" || googleUser.Email == "" || !googleUser.VerifiedEmail {
		return nil, fmt.Errorf("google account is missing required verified identity")
	}
	return &googleUser, nil
}

func (h *AuthHandler) updateGoogleUser(user *models.User, googleUser *GoogleUser) error {
	updates := map[string]interface{}{}
	if user.GoogleID == nil {
		updates["google_id"] = googleUser.ID
	}
	if user.ProfilePicture == nil && googleUser.Picture != "" {
		updates["profile_picture"] = googleUser.Picture
	}
	if len(updates) == 0 {
		return nil
	}
	return h.DB.Model(user).Updates(updates).Error
}

func (h *AuthHandler) findGoogleUser(googleUser *GoogleUser) (*models.User, error) {
	var user models.User
	err := h.DB.Where("google_id = ? OR email = ?", googleUser.ID, strings.ToLower(googleUser.Email)).First(&user).Error
	if err != nil {
		return nil, err
	}
	if err := h.updateGoogleUser(&user, googleUser); err != nil {
		return nil, err
	}
	return &user, nil
}

func (h *AuthHandler) findOrCreateGoogleUser(googleUser *GoogleUser) (*models.User, error) {
	if user, err := h.findGoogleUser(googleUser); err == nil {
		return user, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	baseUsername := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(googleUser.Name), " ", "_"))
	if baseUsername == "" {
		baseUsername = strings.Split(strings.ToLower(googleUser.Email), "@")[0]
	}

	profilePicture := googleUser.Picture
	for attempt := 0; attempt < 5; attempt++ {
		username := baseUsername
		if attempt > 0 {
			username = fmt.Sprintf("%s_%d", baseUsername, time.Now().UnixNano()%1000000000)
		}

		user := models.User{
			Username:       username,
			Email:          strings.ToLower(googleUser.Email),
			GoogleID:       &googleUser.ID,
			ProfilePicture: &profilePicture,
		}
		if err := h.DB.Create(&user).Error; err != nil {
			if !isUniqueViolation(err) {
				return nil, err
			}
			if existing, lookupErr := h.findGoogleUser(googleUser); lookupErr == nil {
				return existing, nil
			}
			continue
		}
		return &user, nil
	}

	return nil, fmt.Errorf("failed to create Google user after concurrent uniqueness conflicts")
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState == "" || cookieState != c.Query("state") {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/auth?error=invalid_state")
		return
	}
	h.setSessionCookie(c, "oauth_state", "", -1)

	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/auth?error=no_code")
		return
	}

	ctx := c.Request.Context()
	oauthToken, err := h.googleOAuthConfig().Exchange(ctx, code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/auth?error=exchange_failed")
		return
	}
	googleUser, err := h.getGoogleUser(ctx, oauthToken)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/auth?error=user_fetch_failed")
		return
	}
	user, err := h.findOrCreateGoogleUser(googleUser)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/auth?error=user_create_failed")
		return
	}
	if err := h.createSession(c, *user); err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/auth?error=session_failed")
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, h.Config.FrontendURL+"/")
}
