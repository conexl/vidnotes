"use client";

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getAccessToken } from '@/lib/auth';

export default function AuthGuard({ children }: { children: React.ReactNode }) {
	const router = useRouter();
	const [ready, setReady] = useState(false);

	useEffect(() => {
		const hasToken = !!getAccessToken();
		if (!hasToken) {
			router.replace('/login');
			return;
		}
		setReady(true);
	}, [router]);

	if (!ready) return null;
	return <>{children}</>;
}
