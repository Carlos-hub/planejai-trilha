import { defineRailway, github, postgres, preserve, project, service } from "railway/iac";

const REPO = "Carlos-hub/planejai-trilha";

// Single source of truth for the project: topology, build settings, and variable
// wiring. Railway's older Config-as-Code (railway.json / railway.toml) is
// deprecated and stops working on 2026-12-01, so nothing lives there.
//
// Both services build from a Dockerfile at the root of their own directory.
export default defineRailway(() => {
  const db = postgres("db");

  const api = service("api", {
    source: github(REPO, { rootDirectory: "backend" }),
    build: { builder: "DOCKERFILE", dockerfilePath: "Dockerfile" },
    healthcheck: "/healthz",
    // restartPolicyType is intentionally left out: ON_FAILURE is Railway's
    // default, and setting it explicitly reads back as null, so every plan
    // would show a change that never converges.
    deploy: { restartPolicyMaxRetries: 5 },
    env: {
      DATABASE_URL: db.env.DATABASE_URL,
      // Pinned so `web` can address the API on a known port over the private
      // network; without it Railway assigns an arbitrary PORT.
      PORT: "8080",
      // Serving over HTTPS, so session cookies get the Secure flag.
      COOKIE_SECURE: "true",
      // Secrets are set once by hand in the Railway dashboard (sealed) and left
      // alone by every apply. They are deliberately NOT generated here:
      // ctx.randomString is a sha256 of its label, so it is deterministic and
      // predictable — fine for a suffix, unusable as a key or session secret.
      TOKEN_ENC_KEY: preserve(),
      SESSION_SECRET: preserve(),
    },
  });

  const web = service("web", {
    source: github(REPO, { rootDirectory: "frontend" }),
    build: { builder: "DOCKERFILE", dockerfilePath: "Dockerfile" },
    healthcheck: "/",
    // restartPolicyType is intentionally left out: ON_FAILURE is Railway's
    // default, and setting it explicitly reads back as null, so every plan
    // would show a change that never converges.
    deploy: { restartPolicyMaxRetries: 5 },
    env: {
      // Railway's own reference syntax, resolved at deploy time. The SDK's
      // api.env.RAILWAY_PRIVATE_DOMAIN returns a reference object that cannot be
      // interpolated into a URL string.
      API_ORIGIN: `http://\${{${api.name}.RAILWAY_PRIVATE_DOMAIN}}:8080`,
    },
  });

  // `api` gets no domain entry: it is reachable only over the private network,
  // and the browser talks to it through the /api proxy in `web`. Railway's
  // generated public domain for `web` is created in the dashboard — generated
  // service domains are not represented in this file.
  return project("planejai", { resources: [db, api, web] });
});
