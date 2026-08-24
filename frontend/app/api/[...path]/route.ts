// Runtime proxy from the Next.js origin to the Go API.
//
// This is deliberately a route handler and not a `rewrites()` entry: Next
// serializes rewrite destinations into the build manifest, so API_ORIGIN would
// be frozen at image-build time instead of read from the deployment's env.
//
// Keeping the browser on a single origin means the session cookie stays
// same-site and the API needs no public domain or CORS headers.

const API_ORIGIN = process.env.API_ORIGIN ?? "http://localhost:8080";

// Hop-by-hop and connection-specific headers must not be forwarded; `host` has
// to be dropped so the upstream sees its own host.
const STRIP_REQUEST = new Set(["host", "connection", "keep-alive", "transfer-encoding"]);
const STRIP_RESPONSE = new Set([
  "connection",
  "keep-alive",
  "transfer-encoding",
  "content-encoding",
  "content-length",
  "set-cookie",
]);

async function proxy(req: Request): Promise<Response> {
  const incoming = new URL(req.url);
  const target = `${API_ORIGIN}${incoming.pathname}${incoming.search}`;

  const headers = new Headers();
  req.headers.forEach((value, key) => {
    if (!STRIP_REQUEST.has(key.toLowerCase())) headers.set(key, value);
  });

  const hasBody = req.method !== "GET" && req.method !== "HEAD";
  let upstream: Response;
  try {
    upstream = await fetch(target, {
      method: req.method,
      headers,
      body: hasBody ? req.body : undefined,
      // Streaming an unbuffered request body requires half-duplex mode.
      ...(hasBody ? { duplex: "half" } : {}),
      redirect: "manual",
      cache: "no-store",
    } as RequestInit);
  } catch {
    return Response.json({ error: "API indisponível" }, { status: 502 });
  }

  const out = new Headers();
  upstream.headers.forEach((value, key) => {
    if (!STRIP_RESPONSE.has(key.toLowerCase())) out.set(key, value);
  });
  // Set-Cookie can repeat; getSetCookie keeps each one separate.
  for (const cookie of upstream.headers.getSetCookie()) out.append("set-cookie", cookie);

  return new Response(upstream.body, { status: upstream.status, headers: out });
}

export const dynamic = "force-dynamic";

export {
  proxy as GET,
  proxy as POST,
  proxy as PUT,
  proxy as PATCH,
  proxy as DELETE,
  proxy as HEAD,
  proxy as OPTIONS,
};
