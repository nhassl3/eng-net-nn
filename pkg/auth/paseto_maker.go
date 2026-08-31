package auth

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
)

type PASETOMaker struct {
	key paseto.V4SymmetricKey
	ttl time.Duration
}

func NewPASETOMaker(keyHex string, ttl time.Duration) (*PASETOMaker, error) {
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("paseto: decode key: %w", err)
	}
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("paseto: key must be exactly 32 bytes, got %d", len(keyBytes))
	}

	key, err := paseto.V4SymmetricKeyFromBytes(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("paseto: create key: %w", err)
	}

	return &PASETOMaker{key: key, ttl: ttl}, nil
}

func (p *PASETOMaker) CreateToken(username, uid, role string) (string, error) {
	return p.createTokenWithJTI(username, uid, role, uuid.New().String(), time.Now())
}

// GetTTL returns TTL for token in seconds
func (p *PASETOMaker) GetTTL() int {
	return int(p.ttl.Seconds())
}

func (p *PASETOMaker) createTokenWithJTI(username, uid, role, jti string, startTime time.Time) (string, error) {
	if startTime.Equal(time.Time{}) {
		startTime = time.Now()
	}

	token := paseto.NewToken()
	token.SetJti(jti)
	token.SetIssuedAt(startTime)
	token.SetExpiration(startTime.Add(p.ttl))
	token.SetString("username", username)
	token.SetString("uid", uid)
	token.SetString("role", role)

	return token.V4Encrypt(p.key, nil), nil
}

func (p *PASETOMaker) CreateRefreshToken(username, uid, role string) (string, *Payload, error) {
	begin := time.Now()
	jti := uuid.New().String()
	token, err := p.createTokenWithJTI(username, uid, role, jti, begin)
	if err != nil {
		return "", nil, err
	}
	return token, &Payload{
		JTI:       jti,
		Username:  username,
		UID:       uid,
		Role:      role,
		IssuedAt:  begin,
		ExpiredAt: begin.Add(p.ttl),
	}, nil
}

func (p *PASETOMaker) VerifyToken(tokenStr string) (*Payload, error) {
	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())

	token, err := parser.ParseV4Local(p.key, tokenStr, nil)
	if err != nil {
		// NotExpired — единственное правило парсера, поэтому RuleError означает
		// именно истёкший токен; всё остальное (ключ, подпись, формат) — битый.
		// Без этого различия транспорт не может ответить 401 вместо 500.
		if errors.Is(err, paseto.RuleError{}) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	jti, err := token.GetJti()
	if err != nil {
		return nil, ErrInvalidToken
	}

	username, err := token.GetString("username")
	if err != nil {
		return nil, ErrInvalidToken
	}

	uid, err := token.GetString("uid")
	if err != nil {
		return nil, ErrInvalidToken
	}

	role, err := token.GetString("role")
	if err != nil {
		return nil, ErrInvalidToken
	}

	issuedAt, err := token.GetIssuedAt()
	if err != nil {
		return nil, ErrInvalidToken
	}

	expiredAt, err := token.GetExpiration()
	if err != nil {
		return nil, ErrExpiredToken
	}

	return &Payload{
		JTI:       jti,
		Username:  username,
		UID:       uid,
		Role:      role,
		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,
	}, nil
}
