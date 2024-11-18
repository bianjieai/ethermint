// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package types

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	regexChainID         = `[a-z]{1,}`
	regexEIP155Separator = `_{1}`
	regexEIP155          = `[1-9][0-9]*`
	regexEpochSeparator  = `-{1}`
	regexEpoch           = `[1-9][0-9]*`
	ethermintChainID     = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)%s(%s)$`,
		regexChainID,
		regexEIP155Separator,
		regexEIP155,
		regexEpochSeparator,
		regexEpoch))
)

// IsValidChainID returns false if the given chain identifier is incorrectly formatted.
func IsValidChainID(chainID string) bool {
	if len(chainID) > 48 {
		return false
	}

	return ethermintChainID.MatchString(chainID)
}

// ParseChainID parses a string chain identifier's epoch to an Ethereum-compatible
// chain-id in *big.Int format. The function returns an error if the chain-id has an invalid format
func ParseChainID(chainID string) (*big.Int, error) {
	if chainIDParser != nil {
		return chainIDParser(chainID)
	}

	chainID = strings.TrimSpace(chainID)
	if len(chainID) > 48 {
		return nil, errorsmod.Wrapf(
			ErrInvalidChainID,
			"chain-id '%s' cannot exceed 48 chars",
			chainID,
		)
	}

	matches := ethermintChainID.FindStringSubmatch(chainID)
	if matches == nil || len(matches) != 4 || matches[1] == "" {
		return nil, errorsmod.Wrapf(ErrInvalidChainID, "%s: %v", chainID, matches)
	}

	// verify that the chain-id entered is a base 10 integer
	chainIDInt, ok := new(big.Int).SetString(matches[2], 10)
	if !ok {
		return nil, errorsmod.Wrapf(
			ErrInvalidChainID,
			"epoch %s must be base-10 integer format",
			matches[2],
		)
	}

	return chainIDInt, nil
}

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
//
// Parameters:
// - valAddr: validator address
// - pubKey: public key
// - selfDelegation: self-delegation amount
// - description: validator description
// - commission: commission rates
// - minSelfDelegation: minimum self-delegation amount
//
// Returns:
// - *stakingtypes.MsgCreateValidator
// - error
func NewMsgCreateValidator(
	valAddr sdk.ValAddress,
	pubKey cryptotypes.PubKey, //nolint:interfacer
	selfDelegation sdk.Coin,
	description stakingtypes.Description,
	commission stakingtypes.CommissionRates,
	minSelfDelegation math.Int,
) (*stakingtypes.MsgCreateValidator, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}
	return &stakingtypes.MsgCreateValidator{
		Description:       description,
		DelegatorAddress:  sdk.AccAddress(valAddr).String(),
		ValidatorAddress:  valAddr.String(),
		Pubkey:            pkAny,
		Value:             selfDelegation,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
	}, nil
}
