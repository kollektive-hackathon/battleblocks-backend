package cosign

type Signable struct {
	Addr        string                 `json:"addr"`
	Args        []interface{}          `json:"args"`
	Cadence     string                 `json:"cadence"`
	Data        map[string]interface{} `json:"data"`
	FType       string                 `json:"f_type"`
	FVsn        string                 `json:"f_vsn"`
	Interaction struct {
		Account struct {
			Addr interface{} `json:"addr"`
		} `json:"account"`
		Accounts map[string]struct {
			Addr  string `json:"addr"`
			KeyID int    `json:"keyId"`
			Kind  string `json:"kind"`
			Role  struct {
				Authorizer bool `json:"authorizer"`
				Param      bool `json:"param"`
				Payer      bool `json:"payer"`
				Proposer   bool `json:"proposer"`
			} `json:"role"`
			SequenceNum interface{} `json:"sequenceNum"`
			Signature   interface{} `json:"signature"`
			TempID      string      `json:"tempId"`
		} `json:"accounts"`
		Arguments      map[string]interface{} `json:"arguments"`
		Assigns        map[string]interface{} `json:"assigns"`
		Authorizations []string               `json:"authorizations"`
		Block          struct {
			Height   interface{} `json:"height"`
			ID       interface{} `json:"id"`
			IsSealed interface{} `json:"isSealed"`
		} `json:"block"`
		Collection struct {
			ID interface{} `json:"id"`
		} `json:"collection"`
		Events struct {
			BlockIDs  []interface{} `json:"blockIds"`
			End       interface{}   `json:"end"`
			EventType interface{}   `json:"eventType"`
			Start     interface{}   `json:"start"`
		} `json:"events"`
		Message struct {
			Arguments      []interface{} `json:"arguments"`
			Authorizations []interface{} `json:"authorizations"`
			Cadence        string        `json:"cadence"`
			ComputeLimit   int           `json:"computeLimit"`
			Params         []interface{} `json:"params"`
			Payer          interface{}   `json:"payer"`
			Proposer       interface{}   `json:"proposer"`
			RefBlock       string        `json:"refBlock"`
		} `json:"message"`
		Params      map[string]interface{} `json:"params"`
		Payer       string                 `json:"payer"`
		Proposer    string                 `json:"proposer"`
		Reason      interface{}            `json:"reason"`
		Status      string                 `json:"status"`
		Tag         string                 `json:"tag"`
		Transaction struct {
			ID interface{} `json:"id"`
		} `json:"transaction"`
	} `json:"interaction"`
	KeyID   int    `json:"keyId"`
	Message string `json:"message"`
	Roles   struct {
		Authorizer bool `json:"authorizer"`
		Param      bool `json:"param"`
		Payer      bool `json:"payer"`
		Proposer   bool `json:"proposer"`
	}
	Voucher Voucher
}

type Voucher struct {
	Arguments    []interface{} `json:"arguments"`
	Authorizers  []string      `json:"authorizers"`
	Cadence      string        `json:"cadence"`
	ComputeLimit int           `json:"computeLimit"`
	EnvelopeSigs []struct {
		Address string `json:"address"`
		KeyID   int    `json:"keyId"`
		Sig     []byte `json:"sig"`
	} `json:"envelopeSigs"`
	Payer       string `json:"payer"`
	PayloadSigs []struct {
		Address string `json:"address"`
		KeyID   int    `json:"keyId"`
		Sig     []byte `json:"sig"`
	} `json:"payloadSigs"`
	ProposalKey struct {
		Address     string      `json:"address"`
		KeyID       int         `json:"keyId"`
		SequenceNum interface{} `json:"sequenceNum"`
	} `json:"proposalKey"`
	RefBlock string `json:"refBlock"`
}
