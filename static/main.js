import Alpine from "https://unpkg.com/alpinejs@3.14.8/dist/module.esm.min.js";
import PromiseWorker from "./worker-promise-client.js";
const promiseWorker = new PromiseWorker(new Worker("worker.js"));

// Debug worker
promiseWorker
  .postMessage(["ping", 10])
  .then((response) => {
    console.log("got response", response);
  })
  .catch((error) => {
    console.log("got error", error);
  });

const placeholderSVG = `
<svg xmlns="http://www.w3.org/2000/svg" width="300" height="150" viewBox="0 0 300 150">
  <rect fill="#ddd" width="300" height="150" />
  <text fill="rgba(0,0,0,0.5)" font-family="sans-serif" font-size="30" dy="10.5" font-weight="bold" x="50%" y="50%" text-anchor="middle">Select image</text>
</svg>`;

function shitpic() {
  return {
    isProcessing: false,
    files: [],
    fileID: 0,
    currentImg: null,
    previousImg: null,
    error: null,
    durationMS: 5_000,
    quality: 75,
    didCopy: null,

    addFile(file) {
      let obj = { id: ++this.fileID, data: file };
      this.files.unshift(obj);
      this.previousImg = this.currentImg;
      this.currentImg = obj;
    },

    async uglify() {
      this.isProcessing = true;
      this.error = null;
      console.log("starting");
      let start = new Date();
      try {
        let output = await promiseWorker.postMessage([
          "uglify",
          [this.currentImg.data, this.durationMS, this.quality],
        ]);
        this.addFile(output);
      } catch (err) {
        this.error = err;
      }
      console.log("done", new Date() - start);
      this.isProcessing = false;
    },
    async change(ev) {
      let file = this.$refs.fileInput.files[0];
      if (!file) {
        this.input = null;
        return;
      }
      let buf = new Uint8Array(await file.arrayBuffer());
      this.addFile(buf);
    },
    asSrc(buf) {
      return URL.createObjectURL(new Blob([buf]));
    },
    get currentSrc() {
      if (!this.currentImg) {
        const encoded = encodeURIComponent(placeholderSVG);
        return `data:image/svg+xml;charset=UTF-8,${encoded}`;
      }
      return this.asSrc(this.currentImg.data);
    },
    get previousSrc() {
      return this.previousImg ? this.asSrc(this.previousImg.data) : "";
    },
    async copyImage(src) {
      let img = new Image();
      img.onload = () => {
        const canvas = document.createElement("canvas");
        const ctx = canvas.getContext("2d");
        canvas.width = img.naturalWidth;
        canvas.height = img.naturalHeight;
        ctx.drawImage(img, 0, 0);
        canvas.toBlob(async (blob) => {
          await navigator.clipboard.write([
            new ClipboardItem({ "image/png": blob }),
          ]);
          clearTimeout(this.didCopy);
          this.didCopy = setTimeout(() => {
            this.didCopy = null;
          }, 1_000);
        });
      };
      img.src = src;
    },
    async pasteImage() {
      try {
        const clipboardContents = await navigator.clipboard.read();
        for (const item of clipboardContents) {
          if (!item.types.includes("image/png")) {
            continue;
          }
          let blob = await item.getType("image/png");
          let file;
          if ("bytes" in blob) {
            file = await blob.bytes();
          } else {
            // Chrome doesn't give us .bytes() for some reason.
            let buf = await blob.arrayBuffer();
            file = new Uint8Array(buf);
          }
          this.addFile(file);
          return;
        }
        this.error = "Clipboard does not contain image data.";
        return;
      } catch (error) {
        this.error = error;
      }
    },
  };
}

Alpine.data("shitpic", shitpic);
Alpine.start();
