package validator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teragrid/dgrid/pkg/crypto/ed25519"
	cmn "github.com/teragrid/dgrid/pkg/common"
)

func TestIsConnTimeoutForFundamentalTimeouts(t *testing.T) {
	// Generate a networking timeout
	dialer := DialTCPFn(testFreeTCPAddr(t), time.Millisecond, ed25519.GenPrivKey())
	_, err := dialer()
	assert.Error(t, err)
	assert.True(t, IsConnTimeout(err))
}

func TestIsConnTimeoutForWrappedConnTimeouts(t *testing.T) {
	dialer := DialTCPFn(testFreeTCPAddr(t), time.Millisecond, ed25519.GenPrivKey())
	_, err := dialer()
	assert.Error(t, err)
	err = cmn.ErrorWrap(ErrConnTimeout, err.Error())
	assert.True(t, IsConnTimeout(err))
}
