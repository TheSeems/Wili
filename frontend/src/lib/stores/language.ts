import { writable } from "svelte/store";
import { locale } from "$lib/i18n";

export type SupportedLanguage = "en" | "ru";

export const supportedLanguages: Record<SupportedLanguage, string> = {
  en: "English",
  ru: "Русский",
};

export const currentLanguage = writable<SupportedLanguage>("en");

export function setLanguage(lang: SupportedLanguage) {
  locale.set(lang);
  currentLanguage.set(lang);
}
