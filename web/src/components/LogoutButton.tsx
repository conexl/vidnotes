"use client";

import { useRouter } from 'next/navigation';
import { clearTokens } from '@/lib/auth';

export default function LogoutButton() {
	const router = useRouter();
	function onLogout() {
		clearTokens();
		router.replace('/login');
	}
	return (
		<button onClick={onLogout} className="rounded-md border px-3 py-1.5 text-sm hover:bg-neutral-50 dark:hover:bg-neutral-800">Выйти</button>
	);
}
