import { copyFileSync, mkdirSync } from "node:fs";
import { resolve } from "node:path";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

function copyOfficialBrand() {
  const src = resolve(__dirname, "../assets");
  const dest = resolve(__dirname, "public");
  mkdirSync(dest, { recursive: true });
  for (const name of ["logo.svg", "icon.svg"] as const) {
    copyFileSync(resolve(src, name), resolve(dest, name));
  }
}

export default defineConfig({
  plugins: [
    react(),
    {
      name: "copy-official-brand",
      buildStart() {
        copyOfficialBrand();
      },
    },
  ],
  base: "./",
});
