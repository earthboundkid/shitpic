function downloadURI(uri, name) {
  var link = document.createElement("a");
  link.download = name;
  link.href = uri;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  delete link;
}

async function doUglify(e) {
  let file = e.currentTarget.files[0];
  let buf = new Uint8Array(await file.arrayBuffer());
  console.log("starting");
  let start = new Date();
  let result = await window.uglify(buf);
  console.log("done", new Date() - start);
  document.querySelector("img").src = URL.createObjectURL(
    new Blob([result.buffer], { type: "image/jpeg" }),
  );
}

document.addEventListener("alpine:init", () => {
  Alpine.magic("download", () => (uri, name) => downloadURI(uri, name));
  Alpine.magic("uglify", () => (e) => doUglify(e));
});
