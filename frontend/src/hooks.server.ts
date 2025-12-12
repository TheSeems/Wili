import type { Handle } from "@sveltejs/kit";

export const handle: Handle = async ({ event, resolve }) => {
  const host = event.url.hostname;
  if (host === "tg.wili.me") {
    const p = event.url.pathname;
    const allow =
      p === "/tgapp" ||
      p.startsWith("/tgapp/") ||
      p.startsWith("/_app/") ||
      p === "/favicon.ico" ||
      p === "/robots.txt" ||
      p === "/sitemap.xml" ||
      p === "/manifest.webmanifest";

    if (!allow) {
      const target = new URL(`https://wili.me${p}${event.url.search}`);
      return Response.redirect(target.toString(), 308);
    }
  }

  return resolve(event);
};

