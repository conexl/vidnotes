type TokenPair = {
	accessToken: string;
	refreshToken: string;
};

const ACCESS_KEY = 'vn_access_token';
const REFRESH_KEY = 'vn_refresh_token';

export function saveTokens(tokens: TokenPair): void {
	if (typeof window === 'undefined') return;
	localStorage.setItem(ACCESS_KEY, tokens.accessToken);
	localStorage.setItem(REFRESH_KEY, tokens.refreshToken);
}

export function getAccessToken(): string | null {
	if (typeof window === 'undefined') return null;
	return localStorage.getItem(ACCESS_KEY);
}

export function getRefreshToken(): string | null {
	if (typeof window === 'undefined') return null;
	return localStorage.getItem(REFRESH_KEY);
}

export function clearTokens(): void {
	if (typeof window === 'undefined') return;
	localStorage.removeItem(ACCESS_KEY);
	localStorage.removeItem(REFRESH_KEY);
}
