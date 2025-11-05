"use client";

import { useState } from 'react';
import { apiChangePassword } from '@/lib/api';

export default function ChangePasswordPage() {
	const [oldPassword, setOldPassword] = useState('');
	const [newPassword, setNewPassword] = useState('');
	const [saving, setSaving] = useState(false);
	const [message, setMessage] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);

	async function onSubmit(e: React.FormEvent) {
		e.preventDefault();
		setSaving(true);
		setMessage(null);
		setError(null);
		try {
			await apiChangePassword({ old_password: oldPassword, new_password: newPassword });
			setMessage('Пароль изменён');
			setOldPassword('');
			setNewPassword('');
		} catch (e: any) {
			setError(e?.message || 'Не удалось изменить пароль');
		} finally {
			setSaving(false);
		}
	}

	return (
		<div className="space-y-6">
			<h1 className="text-2xl font-semibold">Смена пароля</h1>
			{error && <p className="text-sm text-red-600">{error}</p>}
			{message && <p className="text-sm text-green-600">{message}</p>}
			<form onSubmit={onSubmit} className="space-y-4 rounded-xl border bg-white p-4">
				<div>
					<label className="mb-1 block text-sm">Текущий пароль</label>
					<input type="password" value={oldPassword} onChange={e => setOldPassword(e.target.value)} className="w-full rounded-md border px-3 py-2 outline-none focus:ring-2 focus:ring-black/10" />
				</div>
				<div>
					<label className="mb-1 block text-sm">Новый пароль</label>
					<input type="password" value={newPassword} onChange={e => setNewPassword(e.target.value)} className="w-full rounded-md border px-3 py-2 outline-none focus:ring-2 focus:ring-black/10" />
				</div>
				<button disabled={saving} className="rounded-md bg-black px-4 py-2 text-white disabled:opacity-50">{saving ? 'Сохраняем…' : 'Изменить'}</button>
			</form>
		</div>
	);
}
