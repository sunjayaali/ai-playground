package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"

	"auth/internal/auth"
)

const claimsKey = "claims"
const refreshCookieName = "refresh_token"
const accessCookieName = "access_token"

var allowedOrigin = []string{"http://localhost:3000"}

var accessTokenExtractor = extractors.Chain(
	extractors.FromAuthHeader("Bearer"),
	extractors.FromCookie(accessCookieName),
)

type server struct {
	tokens *auth.TokenService
	users  *auth.UserStore
}

func main() {
	addr := flag.String("addr", ":3002", "listen address")
	accessTTL := flag.Duration("access-ttl", 24*time.Hour, "access token lifetime")
	refreshTTL := flag.Duration("refresh-ttl", 7*24*time.Hour, "refresh token lifetime")
	// Defined before Parse so -secret actually registers;
	// the default still reads JWT_SECRET first, so a real
	// env var beats the flag's zero value.
	secret := flag.String("secret", os.Getenv("JWT_SECRET"), "HMAC secret (-secret, JWT_SECRET env, or .env)")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file (%v)", err)
	}

	if *secret == "" {
		*secret = os.Getenv("JWT_SECRET")
	}
	if *secret == "" {
		log.Fatal("JWT secret required: pass -secret, set JWT_SECRET, or add it to .env")
	}

	s := &server{
		tokens: auth.NewTokenService(*secret, *accessTTL, *refreshTTL),
		users:  auth.NewUserStore(),
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigin,
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodOptions},
		AllowHeaders:     []string{fiber.HeaderContentType, fiber.HeaderAuthorization},
		AllowCredentials: true,
	}))
	s.routes(app)

	log.Printf("listening on %s", *addr)
	log.Fatal(app.Listen(*addr))
}

func (s *server) routes(app *fiber.App) {
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	grp := app.Group("/auth")
	grp.Post("/register", s.register)
	grp.Post("/login", s.login)
	grp.Post("/refresh", s.refresh)
	grp.Post("/logout", s.requireAuth, s.logout)
	grp.Get("/me", s.requireAuth, s.me)
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *server) register(c fiber.Ctx) error {
	var creds credentials
	if err := c.Bind().JSON(&creds); err != nil {
		return s.error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	creds.Username = strings.TrimSpace(creds.Username)
	if len(creds.Username) < 3 {
		return s.error(c, fiber.StatusBadRequest, "username must be at least 3 characters")
	}
	if len(creds.Password) < 8 {
		return s.error(c, fiber.StatusBadRequest, "password must be at least 8 characters")
	}

	id, err := s.users.Create(creds.Username, creds.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return s.error(c, fiber.StatusConflict, "username already taken")
		}
		return s.error(c, fiber.StatusInternalServerError, "could not create user")
	}

	return s.tokenResponse(c, id)
}

func (s *server) login(c fiber.Ctx) error {
	var creds credentials
	if err := c.Bind().JSON(&creds); err != nil {
		return s.error(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	// Authenticate returns the matching user; her UUID becomes
	// the token subject, so sessions survive a username change.
	u, ok := s.users.Authenticate(creds.Username, creds.Password)
	if !ok {
		// One message for wrong-user and wrong-password alike:
		// the 401 body must not leak which half failed.
		return s.error(c, fiber.StatusUnauthorized, "invalid username or password")
	}
	return s.tokenResponse(c, u.ID())
}

func (s *server) refresh(c fiber.Ctx) error {
	raw := c.Cookies(refreshCookieName)
	if raw == "" {
		return s.error(c, fiber.StatusBadRequest, "refresh_token required")
	}
	claims, err := s.tokens.Parse(raw, auth.TokenTypeRefresh)
	if err != nil {
		return s.error(c, fiber.StatusUnauthorized, "invalid refresh token")
	}
	if _, ok := s.users.Get(claims.Subject); !ok {
		return s.error(c, fiber.StatusUnauthorized, "user no longer exists")
	}

	return s.tokenResponse(c, claims.Subject)
}

func (s *server) logout(c fiber.Ctx) error {
	s.clearCookie(c, refreshCookieName)
	s.clearCookie(c, accessCookieName)
	return c.JSON(fiber.Map{"ok": true})
}

func (s *server) me(c fiber.Ctx) error {
	claims, ok := c.Locals(claimsKey).(*auth.Claims)
	if !ok {
		return s.error(c, fiber.StatusUnauthorized, "not authenticated")
	}
	u, ok := s.users.Get(claims.Subject)
	if !ok {
		return s.error(c, fiber.StatusUnauthorized, "user no longer exists")
	}
	return c.JSON(fiber.Map{"id": claims.Subject, "username": u.Username()})
}

func (s *server) requireAuth(c fiber.Ctx) error {
	raw, err := accessTokenExtractor.Extract(c)
	if err != nil {
		return s.error(c, fiber.StatusUnauthorized, "missing token")
	}
	claims, err := s.tokens.Parse(raw, auth.TokenTypeAccess)
	if err != nil {
		return s.error(c, fiber.StatusUnauthorized, "invalid or expired token")
	}
	c.Locals(claimsKey, claims)
	return c.Next()
}

func (s *server) tokenResponse(c fiber.Ctx, subject string) error {
	tok, err := s.tokens.Issue(subject)
	if err != nil {
		return s.error(c, fiber.StatusInternalServerError, "could not issue tokens")
	}

	c.Cookie(&fiber.Cookie{
		Name:     accessCookieName,
		Value:    tok.AccessToken,
		Path:     "/",
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   true,
		HTTPOnly: true,
		MaxAge:   400 * 24 * 3600,
	})
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    tok.RefreshToken,
		Path:     "/",
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   true,
		HTTPOnly: true,
		MaxAge:   400 * 24 * 3600,
	})

	return c.JSON(tok)
}

func (s *server) clearCookie(c fiber.Ctx, name string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Path:     "/",
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   true,
		MaxAge:   -1,
	})
}

func (s *server) error(c fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{"error": msg})
}
