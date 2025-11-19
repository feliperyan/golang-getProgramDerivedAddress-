package main

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"filippo.io/edwards25519"
	"github.com/mr-tron/base58"
)

const MaxSeeds = 16
const MaxSeedLength = 32

var pdaMarkerBytes = []byte("ProgramDerivedAddress")

var (
	ErrPointOnCurve  = errors.New("hash landed on curve")
	ErrInvalidBase58 = errors.New("invalid base58 encoding")
)

// Custom error types for better error handling
type ErrMaxSeedsExceeded struct {
	Count int
}

func (e ErrMaxSeedsExceeded) Error() string {
	return fmt.Sprintf("max seeds exceeded: %d (max: %d)", e.Count, MaxSeeds)
}

type ErrSeedTooLong struct {
	Length int
}

func (e ErrSeedTooLong) Error() string {
	return fmt.Sprintf("seed too long: %d bytes (max: %d)", e.Length, MaxSeedLength)
}

// Address represents a Solana address (base58-encoded 32 bytes)
type Address string

// NewAddress creates a new Address from a base58 string, validating it
func NewAddress(addr string) (Address, error) {
	_, err := DecodeAddress(addr)
	if err != nil {
		return "", err
	}
	return Address(addr), nil
}

// ToBytes converts the Address to its 32-byte representation
func (a Address) ToBytes() ([32]byte, error) {
	return DecodeAddress(string(a))
}

// ProgramDerivedAddressInput contains the inputs for PDA generation
type ProgramDerivedAddressInput struct {
	ProgramAddress Address
	Seeds          [][]byte
}

// ProgramDerivedAddressOutput contains the result of PDA generation
type ProgramDerivedAddressOutput struct {
	Address Address
	Bump    uint8
}

// --- Address Logic ---

func AddressFromBytes(bytes [32]byte) string {
	return base58.Encode(bytes[:])
}

func DecodeAddress(addr string) ([32]byte, error) {
	var arr [32]byte
	b, err := base58.Decode(addr)
	if err != nil {
		return arr, ErrInvalidBase58
	}
	if len(b) != 32 {
		return arr, fmt.Errorf("invalid length: %d", len(b))
	}
	copy(arr[:], b)
	return arr, nil
}

// --- PDA Logic ---

// GetProgramDerivedAddress finds a valid PDA and bump seed
func GetProgramDerivedAddress(input ProgramDerivedAddressInput) (ProgramDerivedAddressOutput, error) {
	// Validate seed count (need room for bump seed)
	if len(input.Seeds)+1 > MaxSeeds {
		return ProgramDerivedAddressOutput{}, ErrMaxSeedsExceeded{Count: len(input.Seeds) + 1}
	}

	// Validate seed lengths
	for i, seed := range input.Seeds {
		if len(seed) > MaxSeedLength {
			return ProgramDerivedAddressOutput{}, ErrSeedTooLong{Length: len(seed)}
		}
		_ = i // suppress unused warning if needed
	}

	// Decode program address
	programIdBytes, err := input.ProgramAddress.ToBytes()
	if err != nil {
		return ProgramDerivedAddressOutput{}, err
	}

	// Try bumps from 255 down to 0
	for bump := 255; bump >= 0; bump-- {
		hasher := sha256.New()

		// 1. Write all user-provided seeds
		for _, seed := range input.Seeds {
			hasher.Write(seed)
		}

		// 2. Write the bump seed
		hasher.Write([]byte{uint8(bump)})

		// 3. Write Program ID
		hasher.Write(programIdBytes[:])

		// 4. Write Marker
		hasher.Write(pdaMarkerBytes)

		var digest [32]byte
		copy(digest[:], hasher.Sum(nil))

		// Check if point is on curve (invalid for PDA)
		p := new(edwards25519.Point)
		if _, err := p.SetBytes(digest[:]); err == nil {
			continue // It IS on the curve, invalid PDA, try next bump
		}

		// Valid PDA found
		return ProgramDerivedAddressOutput{
			Address: Address(AddressFromBytes(digest)),
			Bump:    uint8(bump),
		}, nil
	}

	return ProgramDerivedAddressOutput{}, errors.New("no viable bump found")
}

// CreateProgramDerivedAddress creates a PDA with the provided seeds (including bump)
// This does NOT search for a valid bump - it uses the seeds as-is
func CreateProgramDerivedAddress(input ProgramDerivedAddressInput) (Address, error) {
	// Validate seed count
	if len(input.Seeds) > MaxSeeds {
		return "", ErrMaxSeedsExceeded{Count: len(input.Seeds)}
	}

	// Validate seed lengths
	for _, seed := range input.Seeds {
		if len(seed) > MaxSeedLength {
			return "", ErrSeedTooLong{Length: len(seed)}
		}
	}

	// Decode program address
	programIdBytes, err := input.ProgramAddress.ToBytes()
	if err != nil {
		return "", err
	}

	hasher := sha256.New()

	// 1. Write all user-provided seeds (including bump if provided)
	for _, seed := range input.Seeds {
		hasher.Write(seed)
	}

	// 2. Write Program ID
	hasher.Write(programIdBytes[:])

	// 3. Write Marker
	hasher.Write(pdaMarkerBytes)

	var digest [32]byte
	copy(digest[:], hasher.Sum(nil))

	// Check if point is on curve (invalid for PDA)
	p := new(edwards25519.Point)
	if _, err := p.SetBytes(digest[:]); err == nil {
		return "", ErrPointOnCurve
	}

	return Address(AddressFromBytes(digest)), nil
}

func FindPDA(programIdStr string, seeds [][]byte) (string, uint8, error) {
	if len(seeds) > MaxSeeds {
		return "", 0, fmt.Errorf("too many seeds")
	}

	programIdBytes, err := DecodeAddress(programIdStr)
	if err != nil {
		return "", 0, err
	}

	// Try bumps from 255 down to 0
	for bump := 255; bump >= 0; bump-- {
		hasher := sha256.New()

		// 1. Write all user-provided seeds
		for _, seed := range seeds {
			hasher.Write(seed)
		}

		// 2. Write the bump seed
		hasher.Write([]byte{uint8(bump)})

		// 3. Write Program ID
		hasher.Write(programIdBytes[:])

		// 4. Write Marker
		hasher.Write(pdaMarkerBytes)

		var digest [32]byte
		copy(digest[:], hasher.Sum(nil))

		// Check if point is on curve (invalid for PDA)
		p := new(edwards25519.Point)
		if _, err := p.SetBytes(digest[:]); err == nil {
			continue // It IS on the curve, invalid PDA, try next bump
		}

		// Valid PDA found
		return AddressFromBytes(digest), uint8(bump), nil
	}

	return "", 0, errors.New("no viable bump found")
}
