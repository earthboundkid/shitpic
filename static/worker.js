importScripts("wasm_exec.js");
console.log("Worker is running");

// Load the WASM module with Go code.
const go = new Go();
WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
  .then((result) => {
    go.run(result.instance);
    console.log("Worker loaded WASM module");
  })
  .catch((err) => {
    console.error("Worker failed to load WASM module: ", err);
  });

importScripts("worker-promise-server.js");

registerPromiseWorker(async ([action, payload]) => {
  console.log(`Worker got ${action}`);
  switch (action) {
    case "ping":
      console.log("ponging", payload);
      return ["pong", payload];
    case "uglify":
      return uglify(...payload);
    case "resize":
      return resize(...payload);
    default:
      throw `unknown action '${action}'`;
  }
});
