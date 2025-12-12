<script lang="ts">
  interface Props {
    url: string;
    className?: string;
    displayText?: string;
    variant?: "default" | "inline";
  }

  let { url, className = "", displayText = "", variant = "default" }: Props = $props();

  // Extract domain from URL
  let domain = $derived(
    (() => {
      try {
        const urlObj = new URL(url.startsWith("http") ? url : `https://${url}`);
        return urlObj.hostname.replace("www.", "");
      } catch {
        return url
          .replace(/^https?:\/\//, "")
          .replace("www.", "")
          .split("/")[0];
      }
    })()
  );

  // Get favicon URL using Google's favicon service
  let faviconUrl = $derived(`https://www.google.com/s2/favicons?domain=${domain}&sz=32`);

  // Clean URL for display
  let cleanUrl = $derived(url.startsWith("http") ? url : `https://${url}`);

  // Determine what text to display
  let finalDisplayText = $derived(
    (() => {
      if (displayText) return displayText;

      // For Yandex Market URLs, extract product info
      if (domain.includes("market.yandex.ru")) {
        try {
          const urlObj = new URL(cleanUrl);
          const pathParts = urlObj.pathname.split("/");
          const cardIndex = pathParts.indexOf("card");

          if (cardIndex !== -1 && pathParts[cardIndex + 1]) {
            const productName = pathParts[cardIndex + 1] || "Product";
            const productId = pathParts[cardIndex + 2] || "";

            // Clean and format the product name
            const cleanName = productName
              .replace(/-/g, " ")
              .replace(/\b\w/g, (l) => l.toUpperCase());
            const cleanId = productId.replace(/[^0-9]/g, "");

            return cleanId ? `${cleanName} #${cleanId}` : cleanName;
          }
        } catch {}
      }

      // Default to domain
      return domain;
    })()
  );
</script>

<!-- URL Badge using same pattern as Markdown -->
<a
  href={cleanUrl}
  target="_blank"
  rel="noopener noreferrer"
  class="url-badge url-badge-{variant} {className}"
>
  <img
    src={faviconUrl}
    alt="{domain} favicon"
    class="url-badge-favicon"
    onerror={(e: any) => {
      e.target.style.display = "none";
      e.target.nextElementSibling.style.display = "inline";
    }}
  />
  <svg
    class="url-badge-fallback"
    style="display:none;"
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
  >
    <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
    <polyline points="15,3 21,3 21,9"></polyline>
    <line x1="10" y1="14" x2="21" y2="3"></line>
  </svg>
  <span class="url-badge-domain">{finalDisplayText}</span>
  <svg
    class="url-badge-external"
    width="12"
    height="12"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
  >
    <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
    <polyline points="15,3 21,3 21,9"></polyline>
    <line x1="10" y1="14" x2="21" y2="3"></line>
  </svg>
</a>

<style>
  /* URL Badge Styles - matching Markdown component */
  :global(.url-badge) {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    background-color: rgb(243 244 246);
    border: 1px solid rgb(229 231 235);
    border-radius: 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: rgb(55 65 81);
    text-decoration: none;
    transition: all 0.2s ease;
    margin: 0.25rem 0;
    box-sizing: border-box;
    max-width: 100%;
    min-width: 0;
    overflow: hidden;
  }

  /* Inline variant - smaller and more compact */
  :global(.url-badge-inline) {
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
    gap: 0.25rem;
    margin: 0 0.25rem;
    vertical-align: middle;
  }

  /* Ensure proper spacing between multiple inline badges */
  :global(.url-badge-inline + .url-badge-inline) {
    margin-left: 0.5rem;
  }

  :global(.url-badge:hover) {
    background-color: rgb(229 231 235);
    text-decoration: none;
    transform: translateY(-1px);
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  }

  :global(.dark .url-badge) {
    background-color: rgb(31 41 55);
    border-color: rgb(55 65 81);
    color: rgb(209 213 219);
  }

  :global(.dark .url-badge:hover) {
    background-color: rgb(55 65 81);
  }

  :global(.url-badge-favicon) {
    width: 16px;
    height: 16px;
    flex-shrink: 0;
  }

  :global(.url-badge-fallback) {
    width: 16px;
    height: 16px;
    flex-shrink: 0;
    color: rgb(107 114 128);
  }

  :global(.dark .url-badge-fallback) {
    color: rgb(156 163 175);
  }

  :global(.url-badge-domain) {
    max-width: 100%;
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  :global(.url-badge-external) {
    width: 12px;
    height: 12px;
    flex-shrink: 0;
    opacity: 0.6;
  }
</style>
