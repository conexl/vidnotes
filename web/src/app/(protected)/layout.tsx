"use client";

import { AppBar, Box, Button, Container, Toolbar, Typography } from '@mui/material';
import Link from 'next/link';
import AuthGuard from '../../components/AuthGuard';
import LogoutButton from '../../components/LogoutButton';

export default function ProtectedLayout({ children }: { children: React.ReactNode }) {
	return (
		<AuthGuard>
			<AppBar position="sticky" color="transparent" elevation={0}
				sx={{
					backdropFilter: 'saturate(140%) blur(8px)',
					borderBottom: '1px solid rgba(0,255,200,0.14)',
					boxShadow: '0 0 24px rgba(0,255,200,0.08)',
				}}>
				<Toolbar sx={{ gap: 2, maxWidth: 1200, mx: 'auto', width: '100%' }}>
					<Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 800, letterSpacing: 0.5, textShadow: '0 0 12px rgba(0,255,200,0.2)' }}>
						<Link href="/dashboard" style={{ textDecoration: 'none', color: 'inherit' }}>VidNotes</Link>
					</Typography>
					<Button size="small" component={Link} href="/dashboard" sx={{ '&:hover': { color: '#00FFC8' } }}>Dashboard</Button>
					<Button size="small" component={Link} href="/videos" sx={{ '&:hover': { color: '#00FFC8' } }}>Videos</Button>
					<Button size="small" component={Link} href="/ai" sx={{ '&:hover': { color: '#00FFC8' } }}>AI Sessions</Button>
					<Box sx={{ ml: 1 }}>
						<LogoutButton />
					</Box>
				</Toolbar>
			</AppBar>
			<Container maxWidth="lg" sx={{ py: 4 }}>
				{children}
			</Container>
		</AuthGuard>
	);
}
