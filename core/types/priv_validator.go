package types

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/teragrid/dgrid/pkg/crypto"
	"github.com/teragrid/dgrid/pkg/crypto/ed25519"
)

// Validator defines the functionality of a local Dgrid validator
// that signs votes and proposals, and never double signs.
type Validator interface {
	GetPubKey() crypto.PubKey

	SignVote(leagueID string, vote *Vote) error
	SignProposal(leagueID string, proposal *Proposal) error
}

//----------------------------------------
// Misc.

type ValidatorsByAddress []Validator

func (pvs ValidatorsByAddress) Len() int {
	return len(pvs)
}

func (pvs ValidatorsByAddress) Less(i, j int) bool {
	return bytes.Compare(pvs[i].GetPubKey().Address(), pvs[j].GetPubKey().Address()) == -1
}

func (pvs ValidatorsByAddress) Swap(i, j int) {
	it := pvs[i]
	pvs[i] = pvs[j]
	pvs[j] = it
}

//----------------------------------------
// MockPV

// MockPV implements Validator without any safety or persistence.
// Only use it for testing.
type MockPV struct {
	privKey              crypto.PrivKey
	breakProposalSigning bool
	breakVoteSigning     bool
}

func NewMockPV() *MockPV {
	return &MockPV{ed25519.GenPrivKey(), false, false}
}

// NewMockPVWithParams allows one to create a MockPV instance, but with finer
// grained control over the operation of the mock validator. This is useful for
// mocking test failures.
func NewMockPVWithParams(privKey crypto.PrivKey, breakProposalSigning, breakVoteSigning bool) *MockPV {
	return &MockPV{privKey, breakProposalSigning, breakVoteSigning}
}

// Implements Validator.
func (pv *MockPV) GetPubKey() crypto.PubKey {
	return pv.privKey.PubKey()
}

// Implements Validator.
func (pv *MockPV) SignVote(leagueID string, vote *Vote) error {
	useLeagueID := leagueID
	if pv.breakVoteSigning {
		useLeagueID = "incorrect-chain-id"
	}
	signBytes := vote.SignBytes(useLeagueID)
	sig, err := pv.privKey.Sign(signBytes)
	if err != nil {
		return err
	}
	vote.Signature = sig
	return nil
}

// Implements Validator.
func (pv *MockPV) SignProposal(leagueID string, proposal *Proposal) error {
	useLeagueID := leagueID
	if pv.breakProposalSigning {
		useLeagueID = "incorrect-chain-id"
	}
	signBytes := proposal.SignBytes(useLeagueID)
	sig, err := pv.privKey.Sign(signBytes)
	if err != nil {
		return err
	}
	proposal.Signature = sig
	return nil
}

// String returns a string representation of the MockPV.
func (pv *MockPV) String() string {
	addr := pv.GetPubKey().Address()
	return fmt.Sprintf("MockPV{%v}", addr)
}

// XXX: Implement.
func (pv *MockPV) DisableChecks() {
	// Currently this does nothing,
	// as MockPV has no safety checks at all.
}

type erroringMockPV struct {
	*MockPV
}

var ErroringMockPVErr = errors.New("erroringMockPV always returns an error")

// Implements Validator.
func (pv *erroringMockPV) SignVote(leagueID string, vote *Vote) error {
	return ErroringMockPVErr
}

// Implements Validator.
func (pv *erroringMockPV) SignProposal(leagueID string, proposal *Proposal) error {
	return ErroringMockPVErr
}

// NewErroringMockPV returns a MockPV that fails on each signing request. Again, for testing only.
func NewErroringMockPV() *erroringMockPV {
	return &erroringMockPV{&MockPV{ed25519.GenPrivKey(), false, false}}
}
