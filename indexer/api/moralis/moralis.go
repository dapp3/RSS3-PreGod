package moralis

import (
	"fmt"
	"os"

	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/types"
	"github.com/joho/godotenv"
	jsoniter "github.com/json-iterator/go"
)

var jsoni = jsoniter.ConfigCompatibleWithStandardLibrary

func GetMoralisApiKey() string {
	if err := godotenv.Load(".env"); err != nil {
		return ""
	}

	return os.Getenv("MoralisApiKey")
}

func GetNFTs(userAddress string, chainType string, apiKey string) (types.NFTResult, error) {
	var headers = map[string]string{
		"accept":    "application/json",
		"X-API-Key": apiKey,
	}

	// Gets all NFT items of user
	apiUrl := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%s/nft?chain=%s&format=decimal", userAddress, chainType)
	response, _ := Get(apiUrl, headers)

	res := new(types.NFTResult)

	err := jsoni.Unmarshal(response, &res)
	if err != nil {
		return types.NFTResult{}, err
	}

	return *res, nil
}

func GetNFTTransfers(userAddress string, chainType string, apiKey string) (types.NFTTransferResult, error) {
	var headers = map[string]string{
		"accept":    "application/json",
		"X-API-Key": apiKey,
	}

	// Gets all NFT transfers of user
	apiUrl := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%s/nft/transfers?chain=%s&format=decimal&direction=both",
		userAddress, chainType)
	response, _ := Get(apiUrl, headers)

	res := new(types.NFTTransferResult)

	err := jsoni.Unmarshal(response, &res)
	if err != nil {
		return types.NFTTransferResult{}, err
	}

	return *res, nil
}
