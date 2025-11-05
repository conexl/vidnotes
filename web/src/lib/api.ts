import { API_URL } from './config';
import { getAccessToken, getRefreshToken, saveTokens, clearTokens } from './auth';

async function fetchJson<T>(input: string, init: RequestInit = {}): Promise<T> {
	const headers: HeadersInit = {
		'Content-Type': 'application/json',
		...(init.headers || {}),
	};

	const access = getAccessToken();
	if (access) {
		(headers as Record<string, string>).Authorization = `Bearer ${access}`;
	}

	const res = await fetch(input, { ...init, headers, credentials: 'include' });
	if (res.status === 401) {
		const refreshed = await tryRefreshToken();
		if (!refreshed) {
			clearTokens();
			throw new Error('Unauthorized');
		}
		const retryHeaders: HeadersInit = {
			'Content-Type': 'application/json',
			...(init.headers || {}),
		};
		const newAccess = getAccessToken();
		if (newAccess) {
			(retryHeaders as Record<string, string>).Authorization = `Bearer ${newAccess}`;
		}
		const retryRes = await fetch(input, { ...init, headers: retryHeaders, credentials: 'include' });
		if (!retryRes.ok) {
			throw new Error(await safeErrorText(retryRes));
		}
		return (await parseDataEnvelope<T>(retryRes));
	}
	if (!res.ok) {
		throw new Error(await safeErrorText(res));
	}
	return parseDataEnvelope<T>(res);
}

async function parseDataEnvelope<T>(res: Response): Promise<T> {
	const raw = await res.json();
	return (raw && typeof raw === 'object' && 'data' in raw) ? (raw.data as T) : (raw as T);
}

async function safeErrorText(res: Response): Promise<string> {
	try {
		const data = await res.json();
		if (data && typeof data === 'object') {
			return data.error || data.message || res.statusText;
		}
		return res.statusText;
	} catch {
		return res.statusText;
	}
}

async function tryRefreshToken(): Promise<boolean> {
	const refresh = getRefreshToken();
	if (!refresh) return false;
	try {
		const res = await fetch(`${API_URL}/auth/refresh`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			credentials: 'include',
			body: JSON.stringify({ refresh_token: refresh }),
		});
		if (!res.ok) return false;
		const raw = await res.json();
		const data = (raw && typeof raw === 'object' && 'data' in raw) ? raw.data : raw;
		if (data?.access_token && data?.refresh_token) {
			saveTokens({ accessToken: data.access_token, refreshToken: data.refresh_token });
			return true;
		}
		return false;
	} catch {
		return false;
	}
}

// AUTH
export async function apiRegister(input: { email: string; password: string; name?: string }): Promise<void> {
	const data = await fetchJson<{ user_id: string; tokens: { access_token: string; refresh_token: string } }>(
		`${API_URL}/auth/register`,
		{
			method: 'POST',
			body: JSON.stringify(input),
		}
	);
	saveTokens({ accessToken: data.tokens.access_token, refreshToken: data.tokens.refresh_token });
}

export async function apiLogin(input: { email: string; password: string }): Promise<void> {
	const data = await fetchJson<{ user_id: string; tokens: { access_token: string; refresh_token: string } }>(
		`${API_URL}/auth/login`,
		{
			method: 'POST',
			body: JSON.stringify(input),
		}
	);
	saveTokens({ accessToken: data.tokens.access_token, refreshToken: data.tokens.refresh_token });
}

// USER
export async function apiGetProfile(): Promise<any> {
	return fetchJson<any>(`${API_URL}/user/profile`);
}

export async function apiUpdateProfile(input: Record<string, unknown>): Promise<any> {
	return fetchJson<any>(`${API_URL}/user/profile`, { method: 'PUT', body: JSON.stringify(input) });
}

export async function apiChangePassword(input: { old_password: string; new_password: string }): Promise<any> {
	return fetchJson<any>(`${API_URL}/user/change-password`, { method: 'POST', body: JSON.stringify(input) });
}

export async function apiGetAnalytics(): Promise<any> {
	return fetchJson<any>(`${API_URL}/user/analytics`);
}

// VIDEOS
export async function apiUploadVideo(file: File): Promise<any> {
	const form = new FormData();
	form.append('file', file);
	const headers: HeadersInit = {};
	const access = getAccessToken();
	if (access) (headers as Record<string, string>).Authorization = `Bearer ${access}`;
	const res = await fetch(`${API_URL}/videos/upload`, {
		method: 'POST',
		body: form,
		headers,
		credentials: 'include',
	});
	if (res.status === 401) {
		const refreshed = await tryRefreshToken();
		if (!refreshed) throw new Error('Unauthorized');
		const retryHeaders: HeadersInit = {};
		const newAccess = getAccessToken();
		if (newAccess) (retryHeaders as Record<string, string>).Authorization = `Bearer ${newAccess}`;
		const retry = await fetch(`${API_URL}/videos/upload`, { method: 'POST', body: form, headers: retryHeaders, credentials: 'include' });
		if (!retry.ok) throw new Error(await safeErrorText(retry));
		return parseDataEnvelope<any>(retry);
	}
	if (!res.ok) throw new Error(await safeErrorText(res));
	return parseDataEnvelope<any>(res);
}

export async function apiGetUserVideos(): Promise<any> {
	return fetchJson<any>(`${API_URL}/videos/`);
}

export async function apiGetVideoStatus(id: string): Promise<any> {
	return fetchJson<any>(`${API_URL}/videos/${id}`);
}

export async function apiGetVideoResult(id: string): Promise<any> {
	return fetchJson<any>(`${API_URL}/videos/${id}/result`);
}

export async function apiDeleteVideo(id: string): Promise<any> {
	return fetchJson<any>(`${API_URL}/videos/${id}`, { method: 'DELETE' });
}

// AI SESSIONS
export async function apiCreateSession(input: { title?: string; video_id: string }): Promise<any> {
	return fetchJson<any>(`${API_URL}/ai/sessions`, { method: 'POST', body: JSON.stringify(input) });
}

export async function apiGetSessions(): Promise<any[]> {
	return fetchJson<any[]>(`${API_URL}/ai/sessions`);
}

export async function apiGetSession(id: string): Promise<any> {
	return fetchJson<any>(`${API_URL}/ai/sessions/${id}`);
}

export async function apiSendMessage(id: string, input: { content: string }): Promise<any> {
	// backend expects {"message": string}
	return fetchJson<any>(`${API_URL}/ai/sessions/${id}/message`, { method: 'POST', body: JSON.stringify({ message: input.content }) });
}

export async function apiDeleteSession(id: string): Promise<any> {
	return fetchJson<any>(`${API_URL}/ai/sessions/${id}`, { method: 'DELETE' });
}
