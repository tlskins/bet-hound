package betting

import (
	"fmt"

	"bet-hound/cmd/db"
	t "bet-hound/cmd/types"
)

func AcceptBet(user *t.User, betId string, accept bool) (*t.Bet, *t.Notification, error) {
	bet, err := db.FindBetById(betId)
	if err != nil {
		return nil, nil, err
	} else if bet.BetStatus.String() != "Pending Approval" {
		return nil, nil, fmt.Errorf("Cannot accept a bet with status: %s.", bet.BetStatus.String())
	} else if (bet.Recipient != nil && bet.Recipient.Id != user.Id) && bet.Proposer.Id != user.Id {
		return nil, nil, fmt.Errorf("You are not involved with this bet.")
	}

	var status t.BetStatus
	if accept {
		dftFk := "-1"
		if bet.Proposer.Id == user.Id {
			bet.ProposerReplyFk = &dftFk
		} else if bet.Recipient != nil && bet.Recipient.Id == user.Id {
			bet.RecipientReplyFk = &dftFk
		} else if bet.Recipient == nil && user.Id != bet.Proposer.Id {
			bet.Recipient = user.IndexUser()
			bet.RecipientReplyFk = &dftFk
		}

		if bet.ProposerReplyFk != nil && bet.RecipientReplyFk != nil {
			status = t.BetStatusFromString("Accepted")
		}
	} else {
		if bet.Proposer.Id == user.Id {
			status = t.BetStatusFromString("Withdrawn")
		} else if bet.Recipient.Id == user.Id {
			status = t.BetStatusFromString("Declined")
		}
	}
	bet.BetStatus = status
	if err = db.UpsertBet(bet); err != nil {
		return nil, nil, err
	}
	if _, err = TweetBetApproval(bet, nil); err != nil {
		return nil, nil, err
	}
	note, _ := db.SyncBetWithUsers("Update", bet)
	return bet, note, nil
}
