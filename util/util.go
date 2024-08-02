package util

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/argon2"
)

type MainData struct {
	Title string
	Data  interface{}
}

func CreateErrorData(err interface{}) MainData {
	return MainData{
		Title: "Error",
		Data:  err,
	}
}

func RenderError(c echo.Context, err error) error {
	return c.Render(http.StatusInternalServerError, "error", CreateErrorData(err))
}

func CreateData(title string, data interface{}) MainData {
	return MainData{
		Title: title,
		Data:  data,
	}
}

func RenderData(c echo.Context, template string, title string, data interface{}) error {
	return c.Render(http.StatusOK, template, CreateData(title, data))
}

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateHashFromPassword(password string) (string, error) {
	p := params{
		memory:      64 * 1024,
		iterations:  4,
		parallelism: 4,
		saltLength:  16,
		keyLength:   32,
	}

	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash), nil
}

func decodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errors.New("invalid hash")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errors.New("incompatible version")
	}

	p = &params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func ComparePasswordAndHash(password, hash string) (bool, error) {
	p, salt, hashBytes, err := decodeHash(hash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	return subtle.ConstantTimeCompare(hashBytes, otherHash) == 1, nil
}
