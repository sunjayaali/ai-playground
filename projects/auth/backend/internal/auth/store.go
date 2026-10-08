package auth

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserExists = errors.New("user already exists")

// UserStore keeps users in memory. It is a stand-in for a real
// database: everything is lost on restart, which is fine for an
// auth demo but not for production.
type UserStore struct {
	mu        sync.RWMutex
	users     map[string]user   // user id -> user
	usernames map[string]string // username -> user id
}

type user struct {
	id       string // UUID, the identity minted into tokens
	username string
	hash     []byte
}

func NewUserStore() *UserStore {
	return &UserStore{
		users:     make(map[string]user),
		usernames: make(map[string]string),
	}
}

// Create hashes the password before taking the write lock: bcrypt
// is deliberately slow, and holding the lock through it would
// serialize every registration on the CPU. It returns the new
// user's UUID, the identity that gets minted into her tokens.
func (s *UserStore) Create(username, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.usernames[username]; ok {
		return "", ErrUserExists
	}
	id := uuid.NewString()
	s.usernames[username] = id
	s.users[id] = user{id: id, username: username, hash: hash}
	return id, nil
}

// ID returns the user's UUID, the identity minted
// into her tokens.
func (u user) ID() string {
	return u.id
}

// Username returns the user's login name.
func (u user) Username() string {
	return u.username
}

// Get returns the user with the given id — the UUID handed
// out at registration, which is also the JWT subject.
func (s *UserStore) Get(id string) (user, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

// FindByUsername resolves a login name to its user.
func (s *UserStore) FindByUsername(username string) (user, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.usernames[username]
	if !ok {
		return user{}, false
	}
	return s.users[id], true
}

// dummyHash keeps login timing uniform: comparing against it for an
// unknown user costs the same bcrypt work as for a real one.
var dummyHash = func() []byte {
	h, err := bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return h
}()

// Authenticate reports whether username/password match, and
// returns the matching user so the caller gets her id. A missing
// user and a wrong password take the same path — and the same
// time — so the response cannot tell an attacker which half failed.
func (s *UserStore) Authenticate(username, password string) (user, bool) {
	u, ok := s.FindByUsername(username)
	if !ok {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return user{}, false
	}
	if bcrypt.CompareHashAndPassword(u.hash, []byte(password)) != nil {
		return user{}, false
	}
	return u, true
}
