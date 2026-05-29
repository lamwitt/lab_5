package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"books-api/internal/cache"
	"books-api/internal/config"
	"books-api/internal/dto"
	"books-api/internal/models"
	"books-api/internal/repository"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	cfg       *config.Config
	cache     *cache.CacheService
}

func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	cfg *config.Config,
	cache *cache.CacheService,
) *AuthService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, cfg: cfg, cache: cache}
}

func (s *AuthService) Register(d *dto.RegisterDTO) (*models.User, error) {
	existing, err := s.userRepo.FindByEmail(d.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, salt, err := s.hashPassword(d.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{Email: d.Email, PasswordHash: hash, Salt: salt}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(d *dto.LoginDTO) (string, string, error) {
	user, err := s.userRepo.FindByEmail(d.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	if !s.verifyPassword(d.Password, user.PasswordHash, user.Salt) {
		return "", "", ErrInvalidCredentials
	}

	return s.IssueTokenPair(user.ID)
}

func (s *AuthService) Refresh(rawRefreshToken string) (string, string, error) {
	claims, err := s.parseJWT(rawRefreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	stored, err := s.tokenRepo.FindByHash(hashToken(rawRefreshToken))
	if err != nil || stored.Revoked || time.Now().After(stored.ExpiresAt) {
		return "", "", ErrInvalidToken
	}

	_ = s.tokenRepo.RevokeByID(stored.ID)

	return s.IssueTokenPair(claims.UserID)
}

func (s *AuthService) Logout(accessToken, refreshToken string) {
	ctx := context.Background()
	if accessToken != "" {
		if claims, err := s.parseJWT(accessToken, s.cfg.JWTAccessSecret); err == nil {
			_ = s.cache.Del(ctx, jtiKey(claims.UserID, claims.ID))
		}
		_ = s.tokenRepo.RevokeByHash(hashToken(accessToken))
	}
	if refreshToken != "" {
		_ = s.tokenRepo.RevokeByHash(hashToken(refreshToken))
	}
}

func (s *AuthService) LogoutAll(userID uuid.UUID) error {
	ctx := context.Background()
	_ = s.cache.DelByPattern(ctx, fmt.Sprintf("wp:auth:user:%s:jti:*", userID))
	_ = s.cache.Del(ctx, profileKey(userID))
	return s.tokenRepo.RevokeAllByUser(userID)
}

func (s *AuthService) ValidateAccessToken(rawToken string) (*JWTClaims, error) {
	claims, err := s.parseJWT(rawToken, s.cfg.JWTAccessSecret)
	if err != nil {
		return nil, ErrInvalidToken
	}

	ctx := context.Background()
	// Проверяем JTI в Redis (быстрый путь)
	if s.cache.Exists(ctx, jtiKey(claims.UserID, claims.ID)) {
		return claims, nil
	}

	// Фолбэк: проверяем в БД (если Redis недоступен или JTI не найден)
	stored, err := s.tokenRepo.FindByHash(hashToken(rawToken))
	if err != nil || stored.Revoked || time.Now().After(stored.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *AuthService) GetUserByID(id uuid.UUID) (*models.User, error) {
	ctx := context.Background()
	key := profileKey(id)

	if val, err := s.cache.Get(ctx, key); err == nil {
		var user models.User
		if json.Unmarshal([]byte(val), &user) == nil {
			return &user, nil
		}
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(user); err == nil {
		_ = s.cache.Set(ctx, key, string(data), time.Duration(s.cfg.CacheTTL)*time.Second)
	}
	return user, nil
}

func (s *AuthService) ForgotPassword(email string) (string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", nil
	}

	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", err
	}
	raw := hex.EncodeToString(rawBytes)

	rec := &models.Token{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hashToken(raw),
		Type:      models.ResetToken,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	return raw, s.tokenRepo.Create(rec)
}

func (s *AuthService) ResetPassword(resetToken, newPassword string) error {
	stored, err := s.tokenRepo.FindByHash(hashToken(resetToken))
	if err != nil || stored.Revoked || stored.Type != models.ResetToken || time.Now().After(stored.ExpiresAt) {
		return ErrInvalidToken
	}

	hash, salt, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(stored.UserID, hash, salt); err != nil {
		return err
	}

	_ = s.cache.Del(context.Background(), profileKey(stored.UserID))
	_ = s.tokenRepo.RevokeByID(stored.ID)
	return s.tokenRepo.RevokeAllByUser(stored.UserID)
}

// --- OAuth Yandex ---

type yandexUserInfo struct {
	ID           string `json:"id"`
	DefaultEmail string `json:"default_email"`
}

func (s *AuthService) GetYandexAuthURL(state string) string {
	return fmt.Sprintf(
		"https://oauth.yandex.ru/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
		s.cfg.OAuthYandexClientID,
		url.QueryEscape(s.cfg.OAuthYandexCallbackURL),
		state,
	)
}

func (s *AuthService) ExchangeYandexCode(code string) (string, string, error) {
	resp, err := http.PostForm("https://oauth.yandex.ru/token", url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {s.cfg.OAuthYandexClientID},
		"client_secret": {s.cfg.OAuthYandexClientSecret},
		"redirect_uri":  {s.cfg.OAuthYandexCallbackURL},
	})
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", "", err
	}
	if tokenResp.Error != "" || tokenResp.AccessToken == "" {
		return "", "", fmt.Errorf("yandex token error: %s", tokenResp.Error)
	}

	req, _ := http.NewRequest("GET", "https://login.yandex.ru/info?format=json", nil)
	req.Header.Set("Authorization", "OAuth "+tokenResp.AccessToken)

	userResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("yandex user info error: status %d", userResp.StatusCode)
	}

	var info yandexUserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&info); err != nil {
		return "", "", err
	}

	return info.ID, info.DefaultEmail, nil
}

func (s *AuthService) FindOrCreateYandexUser(yandexID, email string) (*models.User, error) {
	if user, err := s.userRepo.FindByYandexID(yandexID); err == nil {
		return user, nil
	}

	if user, err := s.userRepo.FindByEmail(email); err == nil {
		user.YandexID = &yandexID
		_ = s.userRepo.Save(user)
		return user, nil
	}

	user := &models.User{Email: email, YandexID: &yandexID}
	return user, s.userRepo.Create(user)
}

// --- Public helpers ---

func (s *AuthService) IssueTokenPair(userID uuid.UUID) (string, string, error) {
	accessID := uuid.New()
	accessExpiry := time.Now().Add(s.cfg.JWTAccessExpiry)
	accessStr, err := s.signJWT(userID, accessID, accessExpiry, s.cfg.JWTAccessSecret)
	if err != nil {
		return "", "", err
	}

	refreshID := uuid.New()
	refreshExpiry := time.Now().Add(s.cfg.JWTRefreshExpiry)
	refreshStr, err := s.signJWT(userID, refreshID, refreshExpiry, s.cfg.JWTRefreshSecret)
	if err != nil {
		return "", "", err
	}

	now := time.Now()
	if err := s.tokenRepo.Create(&models.Token{
		ID: accessID, UserID: userID, TokenHash: hashToken(accessStr),
		Type: models.AccessToken, ExpiresAt: accessExpiry, CreatedAt: now,
	}); err != nil {
		return "", "", err
	}
	if err := s.tokenRepo.Create(&models.Token{
		ID: refreshID, UserID: userID, TokenHash: hashToken(refreshStr),
		Type: models.RefreshToken, ExpiresAt: refreshExpiry, CreatedAt: now,
	}); err != nil {
		return "", "", err
	}

	// Сохраняем JTI access-токена в Redis для быстрой валидации
	ctx := context.Background()
	ttl := time.Until(accessExpiry)
	_ = s.cache.Set(ctx, jtiKey(userID, accessID.String()), userID.String(), ttl)

	return accessStr, refreshStr, nil
}

// --- Private ---

func (s *AuthService) signJWT(userID, tokenID uuid.UUID, expiry time.Time, secret string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func (s *AuthService) parseJWT(tokenStr, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AuthService) hashPassword(password string) (hash, salt string, err error) {
	saltBytes := make([]byte, 16)
	if _, err = rand.Read(saltBytes); err != nil {
		return
	}
	salt = hex.EncodeToString(saltBytes)
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	return string(hashBytes), salt, nil
}

func (s *AuthService) verifyPassword(password, hash, salt string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+salt)) == nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// GenerateState генерирует случайную строку для OAuth state параметра
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return strings.ToLower(hex.EncodeToString(b)), nil
}

// --- Cache key helpers ---

func jtiKey(userID uuid.UUID, jti string) string {
	return fmt.Sprintf("wp:auth:user:%s:jti:%s", userID, jti)
}

func profileKey(userID uuid.UUID) string {
	return fmt.Sprintf("wp:auth:user:%s:profile", userID)
}

