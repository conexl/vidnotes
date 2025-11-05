"use client";

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiRegister } from '@/lib/api';
import { Alert, Box, Button, Paper, Stack, TextField, Typography } from '@mui/material';

export default function RegisterPage() {
	const router = useRouter();
	const [name, setName] = useState('');
	const [email, setEmail] = useState('');
	const [password, setPassword] = useState('');
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	async function onSubmit(e: React.FormEvent) {
		e.preventDefault();
		setError(null);
		setLoading(true);
		try {
			await apiRegister({ email, password, name });
			router.replace('/dashboard');
		} catch (err: any) {
			setError(err?.message || 'Sign up failed');
		} finally {
			setLoading(false);
		}
	}

	return (
		<Box sx={{ display: 'grid', placeItems: 'center', minHeight: '100dvh', p: 2 }}>
			<Paper variant="outlined" sx={{ p: 4, width: '100%', maxWidth: 420, borderRadius: 3 }}>
				<Stack component="form" spacing={2} onSubmit={onSubmit}>
					<div>
						<Typography variant="h5" fontWeight={600}>Create account</Typography>
						<Typography variant="body2" color="text.secondary">Join VidNotes for free</Typography>
					</div>
					<TextField label="Name" value={name} onChange={e => setName(e.target.value)} fullWidth />
					<TextField label="Email" type="email" value={email} onChange={e => setEmail(e.target.value)} required fullWidth />
					<TextField label="Password" type="password" value={password} onChange={e => setPassword(e.target.value)} required fullWidth />
					{error && <Alert severity="error">{error}</Alert>}
					<Button type="submit" variant="contained" disabled={loading}>{loading ? 'Creating…' : 'Sign up'}</Button>
					<Typography variant="body2" color="text.secondary">
						Already have an account? <Link href="/login">Sign in</Link>
					</Typography>
				</Stack>
			</Paper>
		</Box>
	);
}
