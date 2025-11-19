//go:build js && wasm

package main

import (
	"errors"
	"fmt"
	"syscall/js"
)

// --- Helper to parse inputs safely ---

// parseToBytes takes a JS Value and tries to convert it to []byte.
// It handles Strings and Uint8Arrays.
func parseToBytes(val js.Value) ([]byte, error) {
	if val.Type() == js.TypeString {
		return []byte(val.String()), nil
	}

	if val.Get("constructor").Get("name").String() == "Uint8Array" {
		length := val.Length()
		buf := make([]byte, length)
		js.CopyBytesToGo(buf, val)
		return buf, nil
	}

	return nil, errors.New("seed must be String or Uint8Array")
}

// --- WASM Bridge ---

func getProgramDerivedAddressJS(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return map[string]interface{}{"error": "args: (programId, seedsArray)"}
	}

	progID := args[0].String()
	seedsJS := args[1]

	// Convert JS Array to Go Slice of Bytes
	var seeds [][]byte
	length := seedsJS.Length()

	for i := 0; i < length; i++ {
		b, err := parseToBytes(seedsJS.Index(i))
		if err != nil {
			return map[string]interface{}{"error": fmt.Sprintf("seed %d: %v", i, err)}
		}
		seeds = append(seeds, b)
	}

	addr, bump, err := FindPDA(progID, seeds)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	return map[string]interface{}{
		"address": addr,
		"bump":    bump,
	}
}

func main() {
	// Using select{} is cleaner than channel blocking for WASM
	js.Global().Set("getProgramDerivedAddress", js.FuncOf(getProgramDerivedAddressJS))
	println("PDA WASM Initialized")
	select {}
}
