package secretcipher

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCipher_RoundTrip(t *testing.T) {
	c, err := New("server-secret")
	require.NoError(t, err)

	ciphertext, err := c.Encrypt("sk-test-secret")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(ciphertext, Prefix))
	require.NotContains(t, ciphertext, "sk-test-secret")

	plaintext, err := c.Decrypt(ciphertext)
	require.NoError(t, err)
	require.Equal(t, "sk-test-secret", plaintext)
}

func TestCipher_UsesRandomNonce(t *testing.T) {
	c, err := New("server-secret")
	require.NoError(t, err)

	first, err := c.Encrypt("same-secret")
	require.NoError(t, err)
	second, err := c.Encrypt("same-secret")
	require.NoError(t, err)

	require.NotEqual(t, first, second)
}

func TestCipher_WrongSecretFails(t *testing.T) {
	c, err := New("server-secret")
	require.NoError(t, err)
	ciphertext, err := c.Encrypt("sk-test-secret")
	require.NoError(t, err)

	wrong, err := New("different-secret")
	require.NoError(t, err)
	_, err = wrong.Decrypt(ciphertext)
	require.ErrorIs(t, err, ErrInvalidCiphertext)
}

func TestCipher_RejectsEmptyServerSecret(t *testing.T) {
	_, err := New("  ")
	require.ErrorIs(t, err, ErrEmptySecret)
}

func TestCipher_RejectsUnsupportedFormat(t *testing.T) {
	c, err := New("server-secret")
	require.NoError(t, err)

	_, err = c.Decrypt("v0:aes-gcm:value")
	require.ErrorIs(t, err, ErrUnsupportedFormat)
}

func TestCipher_RejectsMalformedCiphertext(t *testing.T) {
	c, err := New("server-secret")
	require.NoError(t, err)

	_, err = c.Decrypt(Prefix + "not base64")
	require.True(t, errors.Is(err, ErrInvalidCiphertext))
}
