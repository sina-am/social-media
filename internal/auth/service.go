package main

import (
	"fmt"
	"log"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

type AuthService interface {
	Authenticate(username, plainPassword string) (*Account, error)
	Register(*Account) error
	ObtainToken(*Account) (*JWTToken, error)
	VerifyToken(*JWTToken) (*Account, error)
	GetAccountIdFromToken(*JWTToken) (uuid.UUID, error)
	Update(uuid.UUID, *AccountUpdateRequest) error
	GetAccountByID(uuid.UUID) (*Account, error)
	AddFollower(accountId uuid.UUID, followerId uuid.UUID) error
}

type localAuthService struct {
	Storer    Storage
	secretKey []byte
}

func NewLocalAuthService(storer Storage, secretKey string) *localAuthService {
	return &localAuthService{
		Storer:    storer,
		secretKey: []byte(secretKey),
	}
}

func (a *localAuthService) Register(account *Account) error {
	return a.Storer.InsertAccount(account)
}
func (a *localAuthService) Authenticate(username, plainPassword string) (*Account, error) {
	account, err := a.Storer.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if !account.VerifyPassword(plainPassword) {
		return nil, fmt.Errorf("invalid credentials")
	}
	// TODO: Update lastlogin
	return account, nil
}

func (a *localAuthService) Update(accountId uuid.UUID, updateReq *AccountUpdateRequest) error {
	account, err := a.Storer.GetByID(accountId)
	if err != nil {
		return err
	}
	if updateReq.Avatar != "" {
		account.Avatar = updateReq.Avatar
	}
	if updateReq.Email != "" {
		account.Email = updateReq.Email
	}
	if updateReq.Name != "" {
		account.Name = updateReq.Name
	}
	if updateReq.Password != "" {
		account.Password, err = HashPassword(updateReq.Password)
		if err != nil {
			return err
		}
	}

	return a.Storer.Update(account)
}

func (a *localAuthService) AddFollower(accountId uuid.UUID, followerId uuid.UUID) error {
	log.Printf("Fire notification for new follower")
	return a.Storer.InsertAccountFollower(accountId, followerId)
}

func (a *localAuthService) GetAccountByID(accountId uuid.UUID) (*Account, error) {
	return a.Storer.GetByID(accountId)
}

func (a *localAuthService) ObtainToken(account *Account) (*JWTToken, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"account_id": account.ID.String(),
		"expired_at": time.Now().Add(time.Hour * 24).Format(time.RFC822),
	})

	tokenStr, err := token.SignedString(a.secretKey)
	if err != nil {
		return nil, err
	}

	return &JWTToken{
		Token: tokenStr,
		Type:  "bearer",
	}, nil
}

func (a *localAuthService) VerifyToken(t *JWTToken) (*Account, error) {
	accountID, err := a.GetAccountIdFromToken(t)
	if err != nil {
		return nil, err
	}

	return a.Storer.GetByID(accountID)
}

func (a *localAuthService) GetAccountIdFromToken(t *JWTToken) (uuid.UUID, error) {
	claims, err := a.decodeToken(t)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(claims["account_id"].(string))
}

func (a *localAuthService) decodeToken(t *JWTToken) (jwt.MapClaims, error) {
	token, err := jwt.Parse(t.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func IsExpired(t string) error {
	expiredAt, err := time.Parse(time.RFC822, t)
	if err != nil {
		return err
	}
	if expiredAt.After(time.Now()) {
		return fmt.Errorf("expired token: %v", expiredAt)
	}
	return nil
}

type metrics struct {
	newRegister    *prometheus.CounterVec
	newLogin       *prometheus.CounterVec
	loginFauilures *prometheus.CounterVec
}

type monitorAuthService struct {
	next    AuthService
	metrics *metrics
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		newRegister: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "register_total",
				Help: "Number account that register.",
			},
			[]string{"auth"},
		),
		newLogin: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "login_total",
				Help: "Number of login attempts.",
			},
			[]string{"auth"},
		),
		loginFauilures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "login_fauilure_total",
				Help: "Number of failed login attempts.",
			},
			[]string{"auth"},
		),
	}
	reg.MustRegister(m.newRegister)
	reg.MustRegister(m.newLogin)
	reg.MustRegister(m.loginFauilures)
	return m
}

func NewMonitorAuthService(next AuthService, reg prometheus.Registerer) *monitorAuthService {
	return &monitorAuthService{next: next, metrics: newMetrics(reg)}
}

func (a *monitorAuthService) Register(account *Account) error {
	err := a.next.Register(account)
	if err == nil {
		a.metrics.newRegister.With(prometheus.Labels{"auth": "register"}).Inc()
	}
	return err
}

func (a *monitorAuthService) Authenticate(username, plainPassword string) (*Account, error) {
	account, err := a.next.Authenticate(username, plainPassword)
	if err != nil {
		a.metrics.loginFauilures.With(prometheus.Labels{"auth": "failed_login"}).Inc()
	} else {
		a.metrics.newLogin.With(prometheus.Labels{"auth": "success_login"}).Inc()
	}
	return account, err
}

func (a *monitorAuthService) ObtainToken(account *Account) (*JWTToken, error) {
	return a.next.ObtainToken(account)
}

func (a *monitorAuthService) VerifyToken(t *JWTToken) (*Account, error) {
	return a.next.VerifyToken(t)
}

func (a *monitorAuthService) GetAccountIdFromToken(t *JWTToken) (uuid.UUID, error) {
	return a.next.GetAccountIdFromToken(t)
}

func (a *monitorAuthService) Update(accountId uuid.UUID, updateReq *AccountUpdateRequest) error {
	return a.next.Update(accountId, updateReq)
}

func (a *monitorAuthService) AddFollower(accountId uuid.UUID, followerId uuid.UUID) error {
	return a.next.AddFollower(accountId, followerId)
}

func (a *monitorAuthService) GetAccountByID(accountId uuid.UUID) (*Account, error) {
	return a.next.GetAccountByID(accountId)
}
