"use client";

import { useEffect, useState } from 'react';
import { apiGetProfile, apiUpdateProfile } from '@/lib/api';

export default function AccountPage() {
	const [profile, setProfile] = useState<any | null>(null);
	const [name, setName] = useState('');
	const [saving, setSaving] = useState(false);
	const [message, setMessage] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		apiGetProfile()
			.then((p) => {
				setProfile(p);
				setName(p?.name || '');
			})
			.catch((e) => setError(e?.message || 'Не удалось получить профиль'));
	}, []);

	async function onSave(e: React.FormEvent) {
		e.preventDefault();
		setSaving(true);
		setError(null);
		setMessage(null);
		try {
			await apiUpdateProfile({ name });
			setMessage('Сохранено');
		} catch (e: any) {
			setError(e?.message || 'Ошибка сохранения');
		} finally {
			setSaving(false);
		}
	}

	return (
		<div className="space-y-6">
			<h1 className="text-2xl font-semibold">Профиль</h1>
			{error && <p className="text-sm text-red-600">{error}</p>}
			{message && <p className="text-sm text-green-600">{message}</p>}
			{profile && (
				<form onSubmit={onSave} className="space-y-4 rounded-xl border bg-white p-4">
					<div>
						<label className="mb-1 block text-sm">Имя</label>
						<input value={name} onChange={e => setName(e.target.value)} className="w-full rounded-md border px-3 py-2 outline-none focus:ring-2 focus:ring-black/10" />
					</div>
					<button disabled={saving} className="rounded-md bg-black px-4 py-2 text-white disabled:opacity-50">{saving ? 'Сохраняем…' : 'Сохранить'}</button>
				</form>
			)}
		</div>
	);
}
