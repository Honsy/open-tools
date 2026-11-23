import { createSite } from "./server.js";

const host = process.env.SITE_HOST || "127.0.0.1";
const port = Number(process.env.SITE_PORT || "8081");
const site = createSite();
site.listen(port, host, () => console.log(`Kaiwulu site listening on ${host}:${port}`));

for (const signal of ["SIGINT", "SIGTERM"]) {
  process.once(signal, () => site.close());
}
