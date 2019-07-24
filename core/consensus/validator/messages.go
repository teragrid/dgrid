package validator

import (
	amino "github.com/teragrid/dgrid/third_party/amino"
	"github.com/teragrid/dgrid/pkg/crypto"
	"github.com/teragrid/dgrid/core/types"
)

// RemoteSignerMsg is sent between SignerServiceEndpoint and the SignerServiceEndpoint client.
type RemoteSignerMsg interface{}

func RegisterRemoteSignerMsg(cdc *amino.Codec) {
	cdc.RegisterInterface((*RemoteSignerMsg)(nil), nil)
	cdc.RegisterConcrete(&PubKeyRequest{}, "teragrid/remotesigner/PubKeyRequest", nil)
	cdc.RegisterConcrete(&PubKeyResponse{}, "teragrid/remotesigner/PubKeyResponse", nil)
	cdc.RegisterConcrete(&SignVoteRequest{}, "teragrid/remotesigner/SignVoteRequest", nil)
	cdc.RegisterConcrete(&SignedVoteResponse{}, "teragrid/remotesigner/SignedVoteResponse", nil)
	cdc.RegisterConcrete(&SignProposalRequest{}, "teragrid/remotesigner/SignProposalRequest", nil)
	cdc.RegisterConcrete(&SignedProposalResponse{}, "teragrid/remotesigner/SignedProposalResponse", nil)
	cdc.RegisterConcrete(&PingRequest{}, "teragrid/remotesigner/PingRequest", nil)
	cdc.RegisterConcrete(&PingResponse{}, "teragrid/remotesigner/PingResponse", nil)
}

// PubKeyRequest requests the consensus public key from the remote signer.
type PubKeyRequest struct{}

// PubKeyResponse is a PrivValidatorSocket message containing the public key.
type PubKeyResponse struct {
	PubKey crypto.PubKey
	Error  *RemoteSignerError
}

// SignVoteRequest is a PrivValidatorSocket message containing a vote.
type SignVoteRequest struct {
	Vote *types.Vote
}

// SignedVoteResponse is a PrivValidatorSocket message containing a signed vote along with a potenial error message.
type SignedVoteResponse struct {
	Vote  *types.Vote
	Error *RemoteSignerError
}

// SignProposalRequest is a PrivValidatorSocket message containing a Proposal.
type SignProposalRequest struct {
	Proposal *types.Proposal
}

// SignedProposalResponse is a PrivValidatorSocket message containing a proposal response
type SignedProposalResponse struct {
	Proposal *types.Proposal
	Error    *RemoteSignerError
}

// PingRequest is a PrivValidatorSocket message to keep the connection alive.
type PingRequest struct {
}

// PingRequest is a PrivValidatorSocket response to keep the connection alive.
type PingResponse struct {
}
