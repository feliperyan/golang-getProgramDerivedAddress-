// src/worker.ts
import "./wasm_exec.js"; // Side-effect import to load the Go class
import wasmModule from "./main.wasm";

// Define the interface for the expected JSON body
interface PdaRequest {
	programId: string;
	seeds: (string | number[])[]; // JSON arrays are number[], we need to cast to Uint8Array later
}

const go = new Go();
let instance: WebAssembly.Instance | undefined;

async function initWasm() {
	if (!instance) {
		// WebAssembly.instantiate in Cloudflare Workers takes the Module directly
		instance = await WebAssembly.instantiate(wasmModule, go.importObject);
		go.run(instance);
	}
}


export default {
	async fetch(request, env, ctx): Promise<Response> {
		await initWasm();

		if (request.method === "POST") {
			try {
				// Cast the parsed JSON to our interface
				const body = await request.json() as PdaRequest;
				const { programId, seeds } = body;

				if (!programId || !seeds) {
					return new Response("Missing programId or seeds", { status: 400 });
				}

				// Transform JSON arrays into Uint8Arrays for the Go bridge
				const processedSeeds = seeds.map((s) => {
					if (Array.isArray(s)) {
						return new Uint8Array(s);
					}
					return s;
				});

				// Call the global function (now typed in types.d.ts)
				const result = globalThis.getProgramDerivedAddress(programId, processedSeeds);

				if (result.error) {
					return new Response(JSON.stringify(result), {
						status: 400,
						headers: { "Content-Type": "application/json" },
					});
				}

				return new Response(JSON.stringify(result), {
					headers: { "Content-Type": "application/json" },
				});
			} catch (err: any) {
				return new Response(`Server Error: ${err.message}`, { status: 500 });
			}
		}

		return new Response("Send a POST request", { status: 200 });
	},
} satisfies ExportedHandler<Env>;