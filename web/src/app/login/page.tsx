"use client";

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiLogin } from '@/lib/api';
import { Alert, Box, Button, Paper, Stack, TextField, Typography } from '@mui/material';

export default function LoginPage() {
	const router = useRouter();
	const [email, setEmail] = useState('');
	const [password, setPassword] = useState('');
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	async function onSubmit(e: React.FormEvent) {
		e.preventDefault();
		setError(null);
		setLoading(true);
		try {
			await apiLogin({ email, password });
			router.replace('/dashboard');
		} catch (err: any) {
			setError(err?.message || 'Sign in failed');
		} finally {
			setLoading(false);
		}
	}

	return (
		<Box sx={{ display: 'grid', placeItems: 'center', minHeight: '100dvh', p: 2 }}>
			<Paper variant="outlined" sx={{ p: 4, width: '100%', maxWidth: 420, borderRadius: 3 }}>
				<Stack component="form" spacing={2} onSubmit={onSubmit}>
					<div>
						<Typography variant="h5" fontWeight={600}>Sign in</Typography>
						<Typography variant="body2" color="text.secondary">Access your VidNotes account</Typography>
					</div>
					<TextField label="Email" type="email" value={email} onChange={e => setEmail(e.target.value)} required fullWidth />
					<TextField label="Password" type="password" value={password} onChange={e => setPassword(e.target.value)} required fullWidth />
					{error && <Alert severity="error">{error}</Alert>}
					<Button type="submit" variant="contained" disabled={loading}>{loading ? 'Signing in…' : 'Sign in'}</Button>
					<Typography variant="body2" color="text.secondary">
						No account? <Link href="/register">Create one</Link>
					</Typography>
				</Stack>
			</Paper>
		</Box>
	);
}
