package rootmulti

import (
	"github.com/cometbft/cometbft/v2/crypto/merkle"

	storetypes "cosmossdk.io/store/types"
)

// RequireProof returns whether proof is required for the subpath.
func RequireProof(subpath string) bool {
	// XXX: create a better convention.
	// Currently, only when query subpath is "/key", will proof be included in
	// response. If there are some changes about proof building in iavlstore.go,
	// we must change code here to keep consistency with iavlStore#Query.
	return subpath == "/key"
}

//-----------------------------------------------------------------------------

// DefaultProofRuntime returns a default proof runtime for the rootMultiStore.
//
// The default proof runtime registers the commitment op decoder for IAVL and
// SimpleMerkle commitments.
//
// XXX: This should be managed by the rootMultiStore which may want to register
// more proof ops?
func DefaultProofRuntime() (prt *merkle.ProofRuntime) {
	prt = merkle.NewProofRuntime()
	prt.RegisterOpDecoder(storetypes.ProofOpIAVLCommitment, storetypes.CommitmentOpDecoder)
	prt.RegisterOpDecoder(storetypes.ProofOpSimpleMerkleCommitment, storetypes.CommitmentOpDecoder)
	return
}
