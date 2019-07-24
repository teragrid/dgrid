package types

import (
	ttime "github.com/teragrid/dgrid/core/types/time"
)

func MakeCommit(blockID BlockID, height int64, round int,
	voteSet *VoteSet,
	validators []Validator) (*Commit, error) {

	// all sign
	for i := 0; i < len(validators); i++ {
		addr := validators[i].GetPubKey().Address()
		vote := &Vote{
			ValidatorAddress: addr,
			ValidatorIndex:   i,
			Height:           height,
			Round:            round,
			Type:             PrecommitType,
			BlockID:          blockID,
			Timestamp:        ttime.Now(),
		}

		_, err := signAddVote(validators[i], vote, voteSet)
		if err != nil {
			return nil, err
		}
	}

	return voteSet.MakeCommit(), nil
}

func signAddVote(privVal Validator, vote *Vote, voteSet *VoteSet) (signed bool, err error) {
	err = privVal.SignVote(voteSet.LeagueID(), vote)
	if err != nil {
		return false, err
	}
	return voteSet.AddVote(vote)
}
