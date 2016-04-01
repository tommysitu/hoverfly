package backends

import (
	"bytes"
	"encoding/json"
	"github.com/pborman/uuid"
	"golang.org/x/crypto/bcrypt"

	log "github.com/Sirupsen/logrus"
	"github.com/SpectoLabs/hoverfly/cache"
)

type User struct {
	UUID     string `json:"uuid" form:"-"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
	IsAdmin  bool   `json:"is_admin" form:"is_admin"`
}

func (u *User) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(u)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeUser(user []byte) (*User, error) {
	var u *User
	buf := bytes.NewBuffer(user)
	dec := json.NewDecoder(buf)
	err := dec.Decode(&u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func NewAuthBackend(tokenCache, userCache cache.Cache) AuthBackend {
	return AuthBackend{
		TokenCache: tokenCache,
		userCache:  userCache,
	}
}

// UserBucketName - default name for BoltDB bucket that stores user info
const UserBucketName = "authbucket"

// TokenBucketName
const TokenBucketName = "tokenbucket"

// BoltCache - container to implement Cache instance with BoltDB backend for storage
type AuthBackend struct {
	TokenCache cache.Cache
	userCache  cache.Cache
}

func (b *AuthBackend) AddUser(username, password string, admin bool) error {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	u := User{
		UUID:     uuid.New(),
		Username: username,
		Password: string(hashedPassword),
		IsAdmin:  admin,
	}
	bts, err := u.Encode()
	if err != nil {
		logUserError(err, username)
		return err
	}
	err = b.userCache.Set([]byte(username), bts)
	return err
}

func (b *AuthBackend) GetUser(username string) (user *User, err error) {
	userBytes, err := b.userCache.Get([]byte(username))

	if err != nil {
		logUserError(err, username)
		return
	}

	user, err = DecodeUser(userBytes)

	if err != nil {
		logUserError(err, username)
		return
	}

	return
}

func (b *AuthBackend) GetAllUsers() (users []User, err error) {
	values, _ := b.userCache.GetAllValues()
	users = make([]User, len(values), len(values))
	for i, user := range values {
		decodedUser, err := DecodeUser(user)
		users[i] = *decodedUser
		return users, err
	}
	return users, err
}

func logUserError(err error, username string) {
	log.WithFields(log.Fields{
		"error":    err.Error(),
		"username": username,
	})
}