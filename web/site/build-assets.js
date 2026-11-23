import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";

const staticDir = new URL("./static/", import.meta.url);
const outputDir = new URL("fingerprinted/", staticDir);
await mkdir(outputDir, { recursive: true });

const assets = {};
for (const extension of ["css", "js"]) {
  const sourceBytes = await readFile(new URL(`site.${extension}`, staticDir));
  const fingerprint = createHash("sha256").update(sourceBytes).digest("hex").slice(0, 16);
  const filename = `site.${fingerprint}.${extension}`;
  await writeFile(new URL(filename, outputDir), sourceBytes);
  assets[extension] = `/static/fingerprinted/${filename}`;
}
await writeFile(new URL("manifest.json", outputDir), JSON.stringify(assets) + "\n");
