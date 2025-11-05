"use client";

import { AppBar, Box, Button, Container, Paper, Toolbar, Typography } from '@mui/material';
import Link from 'next/link';
import { useEffect, useRef, useState } from 'react';
import { getAccessToken } from '@/lib/auth';
import { useRouter } from 'next/navigation';
import { keyframes } from '@mui/system';

const fadeUp = keyframes`
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
`;

export default function Home() {
	const router = useRouter();
	const [guest, setGuest] = useState<boolean | null>(null);
	const heroRef = useRef<HTMLDivElement>(null);
	useEffect(() => {
		const has = !!getAccessToken();
		if (has) router.replace('/dashboard');
		else setGuest(true);
	}, [router]);
	if (!guest) return null;
	return (
		<Box sx={{ bgcolor: 'background.default', color: 'text.primary', minHeight: '100dvh' }}>
			<AppBar position="sticky" color="transparent" elevation={0}>
				<Toolbar sx={{ maxWidth: 1200, mx: 'auto', width: '100%' }}>
					<Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 700 }}>VidNotes</Typography>
					<Button component={Link} href="/login">Sign in</Button>
					<Button variant="contained" component={Link} href="/register" sx={{ ml: 1 }}>Get started</Button>
				</Toolbar>
			</AppBar>
			<Container maxWidth="lg" sx={{ py: 10 }}>
				<Box display="grid" gridTemplateColumns={{ xs: '1fr', md: '1fr 1fr' }} gap={4} alignItems="center">
					<Box>
						<div ref={heroRef}>
							<Typography variant="h3" fontWeight={700} gutterBottom sx={{ textShadow: '0 0 16px rgba(0,255,200,0.15)', animation: `${fadeUp} 700ms ease-out both` }}>
								Summaries and insights from videos in minutes
							</Typography>
							<Typography variant="h6" color="text.secondary" gutterBottom sx={{ animation: `${fadeUp} 900ms ease-out both` }}>
								Upload videos, get key takeaways, and chat with AI about the content.
							</Typography>
							<Box sx={{ mt: 3, display: 'flex', gap: 2, animation: `${fadeUp} 1100ms ease-out both` }}>
								<Button variant="contained" component={Link} href="/register">Start free</Button>
								<Button variant="outlined" component={Link} href="/login">I have an account</Button>
							</Box>
						</div>
					</Box>
					<Box>
						<Paper variant="outlined" sx={{ p: 3, borderRadius: 3, borderColor: 'primary.main', animation: `${fadeUp} 800ms ease-out both` }}>
							<Box sx={{ aspectRatio: '16/9', bgcolor: '#0b1220', borderRadius: 2, boxShadow: '0 0 24px rgba(124,58,237,0.2) inset, 0 0 32px rgba(0,255,200,0.12)' }} />
							<Typography variant="body2" color="text.secondary" sx={{ mt: 1.5 }}>
								Large files supported, fast processing, and AI chat.
							</Typography>
						</Paper>
					</Box>
				</Box>
				<Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr 1fr' }} gap={2} sx={{ mt: 8 }}>
					{[
						{ title: 'Upload', text: 'Up to 500MB' },
						{ title: 'Analytics', text: 'Results and metrics' },
						{ title: 'AI', text: 'Chat over content' },
					].map((f, idx) => (
						<Paper key={f.title} variant="outlined" sx={{ p: 3, borderRadius: 3, animation: `${fadeUp} ${900 + idx * 120}ms ease-out both` }}>
							<Typography fontWeight={600}>{f.title}</Typography>
							<Typography variant="body2" color="text.secondary">{f.text}</Typography>
						</Paper>
					))}
				</Box>
			</Container>
		</Box>
	);
}
