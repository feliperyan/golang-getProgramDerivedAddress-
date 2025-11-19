package main

import (
	"errors"
	"testing"
)

func TestGetProgramDerivedAddress_Success(t *testing.T) {
	// Test with a valid program address and seeds
	programAddr, err := NewAddress("11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("failed to create program address: %v", err)
	}

	input := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds:          [][]byte{[]byte("test-seed")},
	}

	pda, err := GetProgramDerivedAddress(input)
	if err != nil {
		t.Fatalf("GetProgramDerivedAddress failed: %v", err)
	}

	// Verify that a PDA was returned
	if pda.Address == "" {
		t.Error("expected non-empty address")
	}

	// Verify bump seed is in valid range (0-255)
	if pda.Bump > 255 {
		t.Errorf("bump seed out of range: %d", pda.Bump)
	}

	// Verify the returned address is valid by trying to decode it
	_, err = pda.Address.ToBytes()
	if err != nil {
		t.Errorf("returned address is invalid: %v", err)
	}
}

func TestGetProgramDerivedAddress_Deterministic(t *testing.T) {
	// Test that the same inputs always produce the same output
	programAddr, err := NewAddress("11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("failed to create program address: %v", err)
	}

	input := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds:          [][]byte{[]byte("deterministic-test")},
	}

	pda1, err1 := GetProgramDerivedAddress(input)
	if err1 != nil {
		t.Fatalf("first call failed: %v", err1)
	}

	pda2, err2 := GetProgramDerivedAddress(input)
	if err2 != nil {
		t.Fatalf("second call failed: %v", err2)
	}

	if pda1.Address != pda2.Address {
		t.Errorf("addresses don't match: %s != %s", pda1.Address, pda2.Address)
	}

	if pda1.Bump != pda2.Bump {
		t.Errorf("bump seeds don't match: %d != %d", pda1.Bump, pda2.Bump)
	}
}

func TestGetProgramDerivedAddress_MultipleSeeds(t *testing.T) {
	// Test with multiple seeds
	programAddr, err := NewAddress("11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("failed to create program address: %v", err)
	}

	input := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds: [][]byte{
			[]byte("seed1"),
			[]byte("seed2"),
			[]byte("seed3"),
		},
	}

	pda, err := GetProgramDerivedAddress(input)
	if err != nil {
		t.Fatalf("GetProgramDerivedAddress failed: %v", err)
	}

	if pda.Address == "" {
		t.Error("expected non-empty address")
	}
}

func TestGetProgramDerivedAddress_EmptySeeds(t *testing.T) {
	// Test with no seeds (only bump seed will be added)
	programAddr, err := NewAddress("11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("failed to create program address: %v", err)
	}

	input := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds:          [][]byte{},
	}

	pda, err := GetProgramDerivedAddress(input)
	if err != nil {
		t.Fatalf("GetProgramDerivedAddress failed: %v", err)
	}

	if pda.Address == "" {
		t.Error("expected non-empty address")
	}
}

func TestGetProgramDerivedAddress_MaxSeeds(t *testing.T) {
	// Test with maximum allowed seeds (MaxSeeds - 1, since bump seed will be added)
	programAddr, err := NewAddress("11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("failed to create program address: %v", err)
	}

	// Create MaxSeeds - 1 seeds (leaving room for bump seed)
	seeds := make([][]byte, MaxSeeds-1)
	for i := range seeds {
		seeds[i] = []byte{byte(i)}
	}

	input := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds:          seeds,
	}

	pda, err := GetProgramDerivedAddress(input)
	if err != nil {
		t.Fatalf("GetProgramDerivedAddress failed with max seeds: %v", err)
	}

	if pda.Address == "" {
		t.Error("expected non-empty address")
	}
}

func TestGetProgramDerivedAddress_TooManySeeds(t *testing.T) {
	// Test with too many seeds (MaxSeeds, which will exceed limit when bump is added)
	programAddr, err := NewAddress("11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("failed to create program address: %v", err)
	}

	// Create MaxSeeds seeds (will exceed when bump seed is added)
	seeds := make([][]byte, MaxSeeds)
	for i := range seeds {
		seeds[i] = []byte{byte(i)}
	}

	input := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds:          seeds,
	}

	_, err = GetProgramDerivedAddress(input)
	if err == nil {
		t.Fatal("expected error for too many seeds")
	}

	var maxSeedsErr ErrMaxSeedsExceeded
	if !errors.As(err, &maxSeedsErr) {
		t.Errorf("expected ErrMaxSeedsExceeded, got: %v", err)
	}
}

func TestGetProgramDerivedAddress_SeedTooLong(t *testing.T) {
	// Test with a seed that exceeds MaxSeedLength
	programAddr, err := NewAddress("11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("failed to create program address: %v", err)
	}

	// Create a seed longer than MaxSeedLength
	longSeed := make([]byte, MaxSeedLength+1)

	input := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds:          [][]byte{longSeed},
	}

	_, err = GetProgramDerivedAddress(input)
	if err == nil {
		t.Fatal("expected error for seed too long")
	}

	var seedTooLongErr ErrSeedTooLong
	if !errors.As(err, &seedTooLongErr) {
		t.Errorf("expected ErrSeedTooLong, got: %v", err)
	}
}

func TestGetProgramDerivedAddress_InvalidProgramAddress(t *testing.T) {
	// Test with an invalid program address
	input := ProgramDerivedAddressInput{
		ProgramAddress: Address("invalid-base58-!@#$"),
		Seeds:          [][]byte{[]byte("test")},
	}

	_, err := GetProgramDerivedAddress(input)
	if err == nil {
		t.Fatal("expected error for invalid program address")
	}

	if !errors.Is(err, ErrInvalidBase58) {
		t.Errorf("expected ErrInvalidBase58, got: %v", err)
	}
}

func TestGetProgramDerivedAddress_DifferentSeedsDifferentResults(t *testing.T) {
	// Test that different seeds produce different PDAs
	programAddr, err := NewAddress("11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("failed to create program address: %v", err)
	}

	input1 := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds:          [][]byte{[]byte("seed-a")},
	}

	input2 := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds:          [][]byte{[]byte("seed-b")},
	}

	pda1, err1 := GetProgramDerivedAddress(input1)
	if err1 != nil {
		t.Fatalf("first call failed: %v", err1)
	}

	pda2, err2 := GetProgramDerivedAddress(input2)
	if err2 != nil {
		t.Fatalf("second call failed: %v", err2)
	}

	if pda1.Address == pda2.Address {
		t.Error("different seeds produced the same address")
	}
}

func TestGetProgramDerivedAddress_VerifyBumpSeedWorks(t *testing.T) {
	// Test that the returned bump seed actually produces a valid PDA
	programAddr, err := NewAddress("11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("failed to create program address: %v", err)
	}

	originalSeeds := [][]byte{[]byte("verify-bump")}
	input := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds:          originalSeeds,
	}

	pda, err := GetProgramDerivedAddress(input)
	if err != nil {
		t.Fatalf("GetProgramDerivedAddress failed: %v", err)
	}

	// Now manually create a PDA with the returned bump seed
	seedsWithBump := make([][]byte, len(originalSeeds)+1)
	copy(seedsWithBump, originalSeeds)
	seedsWithBump[len(originalSeeds)] = []byte{pda.Bump}

	verifyInput := ProgramDerivedAddressInput{
		ProgramAddress: programAddr,
		Seeds:          seedsWithBump,
	}

	verifyAddr, err := CreateProgramDerivedAddress(verifyInput)
	if err != nil {
		t.Fatalf("CreateProgramDerivedAddress failed with returned bump: %v", err)
	}

	if verifyAddr != pda.Address {
		t.Errorf("addresses don't match: %s != %s", verifyAddr, pda.Address)
	}
}
