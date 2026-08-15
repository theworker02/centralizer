#!/usr/bin/env node
// Copy canonical assets/ marks into the npm brand package and website/public.
import { copyFileSync, mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const assets = join(root, "assets");
const brand = join(root, "packages", "centralizer-brand");
const websitePublic = join(root, "website", "public");

mkdirSync(brand, { recursive: true });
mkdirSync(websitePublic, { recursive: true });

const all = [
  "logo.svg",
  "logo-dark.svg",
  "logo-light.svg",
  "icon.svg",
  "icon-256.png",
  "icon-512.png",
  "github-social-preview.png",
];
const site = ["logo.svg", "icon.svg"];

for (const name of all) {
  copyFileSync(join(assets, name), join(brand, name));
}
for (const name of site) {
  copyFileSync(join(assets, name), join(websitePublic, name));
}

console.log("synced brand assets → packages/centralizer-brand and website/public");
