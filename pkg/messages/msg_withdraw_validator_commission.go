package messages

import (
	cosmosDistributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/gogo/protobuf/proto"
	configTypes "main/pkg/config/types"
	dataFetcher "main/pkg/data_fetcher"
	"main/pkg/types"
	"main/pkg/utils"
)

type MsgWithdrawValidatorCommission struct {
	ValidatorAddress configTypes.Link
	Height           int64
	Amount           []*types.Amount
}

func ParseMsgWithdrawValidatorCommission(data []byte, chain *configTypes.Chain, height int64) (types.Message, error) {
	var parsedMessage cosmosDistributionTypes.MsgWithdrawValidatorCommission
	if err := proto.Unmarshal(data, &parsedMessage); err != nil {
		return nil, err
	}

	return &MsgWithdrawValidatorCommission{
		ValidatorAddress: chain.GetValidatorLink(parsedMessage.ValidatorAddress),
		Height:           height,
	}, nil
}

func (m MsgWithdrawValidatorCommission) Type() string {
	return "MsgWithdrawValidatorCommission"
}

func (m *MsgWithdrawValidatorCommission) GetAdditionalData(fetcher dataFetcher.DataFetcher) {
	rewards, found := fetcher.GetCommissionAtBlock(
		m.ValidatorAddress.Value,
		m.Height,
	)
	if found {
		m.Amount = make([]*types.Amount, len(rewards))

		for index, reward := range rewards {
			m.Amount[index] = &types.Amount{
				Value: utils.StrToFloat64(reward.Amount),
				Denom: reward.Denom,
			}
		}

		price, priceFound := fetcher.GetPrice()
		if priceFound {
			for _, amount := range m.Amount {
				if amount.Denom != fetcher.Chain.BaseDenom {
					continue
				}

				amount.Value /= float64(fetcher.Chain.DenomCoefficient)
				amount.Denom = fetcher.Chain.DisplayDenom
				amount.PriceUSD = amount.Value * price
			}
		}
	}

	if validator, found := fetcher.GetValidator(m.ValidatorAddress.Value); found {
		m.ValidatorAddress.Title = validator.Description.Moniker
	}
}

func (m *MsgWithdrawValidatorCommission) GetValues() map[string]string {
	return map[string]string{
		"type":      "MsgWithdrawValidatorCommission",
		"validator": m.ValidatorAddress.Value,
	}
}
