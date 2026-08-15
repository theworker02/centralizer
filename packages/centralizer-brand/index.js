import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(fileURLToPath(import.meta.url));

/** Directory containing the official marks (copied from repo `assets/`). */
export const assetsDir = root;

export const logo = join(root, "logo.svg");
export const logoDark = join(root, "logo-dark.svg");
export const logoLight = join(root, "logo-light.svg");
export const icon = join(root, "icon.svg");
export const icon256 = join(root, "icon-256.png");
export const icon512 = join(root, "icon-512.png");
export const socialPreview = join(root, "github-social-preview.png");

export default {
  assetsDir,
  logo,
  logoDark,
  logoLight,
  icon,
  icon256,
  icon512,
  socialPreview,
};
