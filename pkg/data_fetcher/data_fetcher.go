package data_fetcher

import (
	"github.com/rs/zerolog"
	"main/pkg/cache"
	priceFetchers "main/pkg/price_fetchers"
	"main/pkg/tendermint/api"
	"main/pkg/types/chains"
	"main/pkg/types/responses"
)

type DataFetcher struct {
	Logger               zerolog.Logger
	Cache                *cache.Cache
	Chain                *chains.Chain
	PriceFetcher         priceFetchers.PriceFetcher
	TendermintApiClients []*api.TendermintApiClient
}

func NewDataFetcher(logger *zerolog.Logger, chain *chains.Chain) *DataFetcher {
	tendermintApiClients := make([]*api.TendermintApiClient, len(chain.APINodes))
	for index, node := range chain.APINodes {
		tendermintApiClients[index] = api.NewTendermintApiClient(logger, node, chain)
	}

	return &DataFetcher{
		Logger: logger.With().
			Str("component", "data_fetcher").
			Str("chain", chain.Name).
			Logger(),
		Cache:                cache.NewCache(logger),
		PriceFetcher:         priceFetchers.GetPriceFetcher(logger, chain),
		Chain:                chain,
		TendermintApiClients: tendermintApiClients,
	}
}

func (f *DataFetcher) GetPrice() (float64, bool) {
	if f.PriceFetcher == nil {
		return 0, false
	}

	if cachedPrice, cachedPricePresent := f.Cache.Get(f.Chain.Name + "_price"); cachedPricePresent {
		if cachedPriceFloat, ok := cachedPrice.(float64); ok {
			return cachedPriceFloat, true
		}

		f.Logger.Error().Msg("Could not convert cached price to float64")
		return 0, false
	}

	notCachedPrice, err := f.PriceFetcher.GetPrice()
	if err != nil {
		f.Logger.Error().Msg("Error fetching price")
		return 0, false
	}

	f.Cache.Set(f.Chain.Name+"_price", notCachedPrice)
	return notCachedPrice, true
}

func (f *DataFetcher) GetValidator(address string) (*responses.Validator, bool) {
	keyName := f.Chain.Name + "_validator_" + address

	if cachedValidator, cachedValidatorPresent := f.Cache.Get(keyName); cachedValidatorPresent {
		if cachedValidatorParsed, ok := cachedValidator.(*responses.Validator); ok {
			return cachedValidatorParsed, true
		}

		f.Logger.Error().Msg("Could not convert cached price to *stakingTypes.Validator")
		return nil, false
	}

	for _, node := range f.TendermintApiClients {
		notCachedValidator, err := node.GetValidator(address)
		if err != nil {
			f.Logger.Error().Msg("Error fetching validator")
			return nil, false
		}

		f.Cache.Set(keyName, notCachedValidator)
		return notCachedValidator, true
	}

	f.Logger.Error().Msg("Could not connect to any nodes to get a validator")
	return nil, true
}
