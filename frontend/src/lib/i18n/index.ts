import { browser } from "$app/environment";
import { init, register, locale, isLoading } from "svelte-i18n";

const FALLBACK_LOCALE = "ru";

register("en", () => import("./locales/en.json"));
register("ru", () => import("./locales/ru.json"));

function detectBrowserLocale(): string | null {
  if (!browser) return null;
  const map = (lang: string | undefined | null): string | null => {
    if (!lang) return null;
    const ll = lang.toLowerCase();
    if (ll.startsWith("ru")) return "ru";
    if (ll.startsWith("en")) return "en";
    return null;
  };
  const langs = (navigator.languages || []).map(map).filter(Boolean) as string[];
  if (langs.length > 0) return langs[0];
  const single = map(navigator.language);
  return single;
}

// Initialize i18n with proper loading
const initI18n = async () => {
  const stored = browser ? window.localStorage.getItem("locale") : null;
  const detected = detectBrowserLocale();
  const initial = stored || detected || FALLBACK_LOCALE;
  await init({
    fallbackLocale: FALLBACK_LOCALE,
    initialLocale: initial,
    loadingDelay: 200,
  });
};

// Initialize immediately
initI18n();

// Save locale preference to localStorage when it changes
if (browser) {
  locale.subscribe((value) => {
    if (value) {
      localStorage.setItem("locale", value);
    }
  });
}

export { locale, isLoading };
