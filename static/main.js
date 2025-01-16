function downloadURI(uri, name) {
  var link = document.createElement("a");
  link.download = name;
  link.href = uri;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  delete link;
}

const placeholderSVG = `
<svg xmlns="http://www.w3.org/2000/svg" width="300" height="150" viewBox="0 0 300 150">
  <rect fill="#ddd" width="300" height="150" />
  <text fill="rgba(0,0,0,0.5)" font-family="sans-serif" font-size="30" dy="10.5" font-weight="bold" x="50%" y="50%" text-anchor="middle">Select image</text>
</svg>`;

function shitpic() {
  return {
    isProcessing: false,
    input: null,
    output: null,
    error: null,

    async process(file) {
      this.isProcessing = true;
      this.output = null;
      console.log("starting");
      let start = new Date();
      try {
        this.output = await window.uglify(this.input);
      } catch (e) {
        console.log("err", e);
        this.error = e;
        this.input = null;
      }
      console.log("done", new Date() - start);
      this.isProcessing = false;
    },
    async change(ev) {
      let file = this.$refs.fileInput.files[0];
      if (!file) {
        this.input = null;
        this.output = null;
        return;
      }
      let buf = new Uint8Array(await file.arrayBuffer());
      this.input = buf;
      await this.process();
    },
    get src() {
      if (!this.output) {
        const encoded = encodeURIComponent(placeholderSVG);
        return `data:image/svg+xml;charset=UTF-8,${encoded}`;
      }
      return URL.createObjectURL(
        new Blob([this.output.buffer], { type: "image/jpeg" }),
      );
    },
    downloadLink: {
      [":download"]() {
        return this.output ? "shitpic.jpeg" : null;
      },
      ["@click"]() {
        if (this.output) {
          return;
        }
        this.$refs.fileInput.click();
      },
      [":href"]() {
        return this.output ? this.src : null;
      },
    },
  };
}

document.addEventListener("alpine:init", () => {
  Alpine.data("shitpic", shitpic);
});
