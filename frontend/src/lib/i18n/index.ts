import { browser } from '$app/environment';
import { init, register, locale, isLoading } from 'svelte-i18n';

const defaultLocale = 'en';

register('en', () => import('./locales/en.json'));
register('ru', () => import('./locales/ru.json'));

// Initialize i18n with proper loading
const initI18n = async () => {
	await init({
		fallbackLocale: defaultLocale,
		initialLocale: browser ? window.localStorage.getItem('locale') || defaultLocale : defaultLocale,
		loadingDelay: 200,
	});
};

// Initialize immediately
initI18n();

// Save locale preference to localStorage when it changes
if (browser) {
	locale.subscribe((value) => {
		if (value) {
			localStorage.setItem('locale', value);
		}
	});
}

export { locale, isLoading };