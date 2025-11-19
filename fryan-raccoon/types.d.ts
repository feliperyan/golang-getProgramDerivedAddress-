// src/types.d.ts

// 1. Fix "Cannot find module './main.wasm'"
// This tells TS that .wasm files export a WebAssembly.Module
declare module "*.wasm" {
  const content: WebAssembly.Module;
  export default content;
}

// 2. Fix "Cannot find name 'Go'"
// This defines the shape of the class provided by wasm_exec.js
declare class Go {
  constructor();
  argv: string[];
  env: { [key: string]: string };
  exit: (code: number) => void;
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
}

// 3. Define your custom Go function on the Global scope
// This allows you to call window.getProgramDerivedAddress() without TS errors
declare global {
  function getProgramDerivedAddress(
    programId: string, 
    seeds: (string | Uint8Array)[]
  ): { address: string; bump: number; error?: string };
}

export {};