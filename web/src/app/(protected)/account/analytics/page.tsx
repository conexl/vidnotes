"use client";

import { useEffect, useState } from 'react';
import { apiGetAnalytics } from '@/lib/api';

export default function AnalyticsPage() {
	const [data, setData] = useState<any | null>(null);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		apiGetAnalytics().then(setData).catch((e) => setError(e?.message || 'Не удалось загрузить аналитику'));
	}, []);

	return (
		<div className="space-y-6">
			<h1 className="text-2xl font-semibold">Аналитика</h1>
			{error && <p className="text-sm text-red-600">{error}</p>}
			{data && (
				<div className="grid gap-4 sm:grid-cols-3">
					<div className="rounded-xl border bg-white p-4">
						<div className="text-sm text-neutral-500">Всего видео</div>
						<div className="text-2xl font-semibold">{data.total_videos ?? '-'}</div>
					</div>
					<div className="rounded-xl border bg-white p-4">
						<div className="text-sm text-neutral-500">Обработано</div>
						<div className="text-2xl font-semibold">{data.processed_videos ?? '-'}</div>
					</div>
					<div className="rounded-xl border bg-white p-4">
						<div className="text-sm text-neutral-500">AI сообщений</div>
						<div className="text-2xl font-semibold">{data.ai_messages ?? '-'}</div>
					</div>
				</div>
			)}
		</div>
	);
}
