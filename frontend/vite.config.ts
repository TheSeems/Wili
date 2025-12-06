import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [sveltekit(), tailwindcss()],
  server: {
    port: 5173,
    allowedHosts: ["ppujm-95-24-109-146.a.free.pinggy.link"],
  },
});
