"use client";

import { useEffect, useState } from 'react';

const STORAGE_KEY = 'vn_theme';

type Theme = 'light' | 'dark';

export default function ThemeToggle() {
	const [theme, setTheme] = useState<Theme>('light');

	useEffect(() => {
		const saved = (typeof window !== 'undefined' ? (localStorage.getItem(STORAGE_KEY) as Theme | null) : null) || 'light';
		applyTheme(saved);
		setTheme(saved);
	}, []);

	function applyTheme(next: Theme) {
		const root = document.documentElement;
		if (next === 'dark') root.classList.add('dark');
		else root.classList.remove('dark');
		localStorage.setItem(STORAGE_KEY, next);
	}

	function toggle() {
		const next: Theme = theme === 'light' ? 'dark' : 'light';
		setTheme(next);
		applyTheme(next);
	}

	return (
		<button onClick={toggle} className="rounded-md border px-3 py-1.5 text-sm hover:bg-neutral-50 dark:hover:bg-neutral-800">
			{theme === 'light' ? '🌙 Тёмная' : '☀️ Светлая'}
		</button>
	);
}
