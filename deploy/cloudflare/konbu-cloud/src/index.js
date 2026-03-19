const UPSTREAM_ORIGIN = "https://konbu-589959123126.asia-northeast1.run.app";

export default {
  async fetch(request) {
    const url = new URL(request.url);

    if (url.pathname === "/__edge-health") {
      return new Response("ok", {
        headers: {
          "content-type": "text/plain; charset=utf-8",
          "x-konbu-cloud-edge": "worker",
        },
      });
    }

    const upstreamUrl = new URL(url.pathname + url.search, UPSTREAM_ORIGIN);
    const upstreamRequest = new Request(upstreamUrl.toString(), request);
    upstreamRequest.headers.set("host", new URL(UPSTREAM_ORIGIN).host);
    upstreamRequest.headers.set("x-forwarded-host", url.host);
    upstreamRequest.headers.set("x-forwarded-proto", url.protocol.replace(":", ""));

    const response = await fetch(upstreamRequest, {
      redirect: "manual",
    });

    const proxiedResponse = new Response(response.body, response);
    proxiedResponse.headers.set("x-konbu-cloud-edge", "worker");
    return proxiedResponse;
  },
};
