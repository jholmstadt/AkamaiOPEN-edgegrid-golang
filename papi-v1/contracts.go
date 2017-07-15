package papi

import (
	"fmt"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/client-v1"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/jsonhooks-v1"
)

// Contracts represents a collection of property manager contracts
type Contracts struct {
	client.Resource
	AccountID string `json:"accountId"`
	Contracts struct {
		Items []*Contract `json:"items"`
	} `json:"contracts"`
}

// NewContracts creates a new Contracts
func NewContracts() *Contracts {
	contracts := &Contracts{}
	contracts.Init()
	return contracts
}

// PostUnmarshalJSON is called after JSON unmarshaling into EdgeHostnames
//
// See: jsonhooks-v1/jsonhooks.Unmarshal()
func (contracts *Contracts) PostUnmarshalJSON() error {
	contracts.Init()

	for key, contract := range contracts.Contracts.Items {
		contracts.Contracts.Items[key].parent = contracts

		if contract, ok := jsonhooks.ImplementsPostJSONUnmarshaler(contract); ok {
			if err := contract.(jsonhooks.PostJSONUnmarshaler).PostUnmarshalJSON(); err != nil {
				return err
			}
		}
	}
	contracts.Complete <- true

	return nil
}

// GetContracts populates Contracts with contract data
//
// API Docs: https://developer.akamai.com/api/luna/papi/resources.html#listcontracts
// Endpoint: GET /papi/v0/contracts
func (contracts *Contracts) GetContracts() error {
	req, err := client.NewRequest(
		Config,
		"GET",
		"/papi/v0/contracts",
		nil,
	)
	if err != nil {
		return err
	}

	res, err := client.Do(Config, req)
	if err != nil {
		return err
	}

	if client.IsError(res) {
		return client.NewAPIError(res)
	}

	if err = client.BodyJSON(res, contracts); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return nil
}

// Contract represents a property contract resource
type Contract struct {
	client.Resource
	parent           *Contracts
	ContractID       string `json:"contractId"`
	ContractTypeName string `json:"contractTypeName"`
}

// NewContract creates a new Contract
func NewContract(parent *Contracts) *Contract {
	contract := &Contract{
		parent: parent,
	}
	contract.Init()
	return contract
}

// GetContract populates a Contract
func (contract *Contract) GetContract() error {
	contracts, err := GetContracts()
	if err != nil {
		return err
	}

	for _, c := range contracts.Contracts.Items {
		if c.ContractID == contract.ContractID {
			contract.parent = c.parent
			contract.ContractTypeName = c.ContractTypeName
			contract.Complete <- true
			return nil
		}
	}
	contract.Complete <- false
	return fmt.Errorf("contract \"%s\" not found", contract.ContractID)
}

// GetProducts gets products associated with a contract
func (contract *Contract) GetProducts() (*Products, error) {
	req, err := client.NewRequest(
		Config,
		"GET",
		fmt.Sprintf(
			"/papi/v0/products?contractId=%s",
			contract.ContractID,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(Config, req)
	if err != nil {
		return nil, err
	}

	if client.IsError(res) {
		return nil, client.NewAPIError(res)
	}

	products := NewProducts()
	if err = client.BodyJSON(res, products); err != nil {
		return nil, err
	}

	return products, nil
}