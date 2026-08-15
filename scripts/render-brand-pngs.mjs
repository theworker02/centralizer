import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { Resvg } from "@resvg/resvg-js";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const assets = join(root, "assets");

function render(svgName, pngName, width) {
  const svg = readFileSync(join(assets, svgName));
  const png = new Resvg(svg, {
    fitTo: { mode: "width", value: width },
    font: { loadSystemFonts: true },
    background: "rgba(0,0,0,0)",
  }).render().asPng();
  writeFileSync(join(assets, pngName), png);
  console.log("wrote", pngName, png.length, "bytes");
}

render("logo.svg", "icon-256.png", 256);
render("logo.svg", "icon-512.png", 512);
render("github-social-preview.svg", "github-social-preview.png", 1280);
