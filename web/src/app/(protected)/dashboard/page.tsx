"use client";

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiGetProfile, apiGetAnalytics, apiGetUserVideos, apiGetSessions } from '@/lib/api';
import { Box, Button, Card, CardContent, CardHeader, Divider, List, ListItem, ListItemText, Skeleton, Stack, Typography } from '@mui/material';

export default function DashboardPage() {
	const [profile, setProfile] = useState<any | null>(null);
	const [analytics, setAnalytics] = useState<any | null>(null);
	const [videos, setVideos] = useState<any[]>([]);
	const [sessions, setSessions] = useState<any[]>([]);
	const [error, setError] = useState<string | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		async function load() {
			try {
				const [p, a, v, s] = await Promise.all([
					apiGetProfile().catch(() => null),
					apiGetAnalytics().catch(() => null),
					apiGetUserVideos().catch(() => []),
					apiGetSessions().catch(() => []),
				]);
				setProfile(p);
				setAnalytics(a);
				setVideos(Array.isArray(v) ? v.slice(0, 5) : []);
				setSessions(Array.isArray(s) ? s.slice(0, 5) : []);
			} catch (e: any) {
				setError(e?.message || 'Failed to load dashboard');
			} finally {
				setLoading(false);
			}
		}
		load();
	}, []);

	return (
		<Stack spacing={3}>
			<Box display="flex" alignItems="end" justifyContent="space-between" gap={2}>
				<div>
					<Typography variant="h5" fontWeight={600}>Dashboard</Typography>
					<Typography variant="body2" color="text.secondary">Welcome{profile?.name ? `, ${profile.name}` : ''}.</Typography>
				</div>
				<Box display="flex" gap={1}>
					<Button component={Link} href="/videos" variant="outlined">Upload video</Button>
					<Button component={Link} href="/ai" variant="contained">New AI session</Button>
				</Box>
			</Box>

			{error && <Typography color="error" variant="body2">{error}</Typography>}

			<Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr 1fr' }} gap={2}>
				{loading ? (
					[...Array(3)].map((_, i) => (
						<Box key={i}><Skeleton variant="rounded" height={112} /></Box>
					))
				) : (
					<>
						<Card>
							<CardContent>
								<Typography variant="body2" color="text.secondary">Total videos</Typography>
								<Typography variant="h5">{analytics?.total_videos ?? videos.length}</Typography>
							</CardContent>
						</Card>
						<Card>
							<CardContent>
								<Typography variant="body2" color="text.secondary">Processed</Typography>
								<Typography variant="h5">{analytics?.processed_videos ?? '-'}</Typography>
							</CardContent>
						</Card>
						<Card>
							<CardContent>
								<Typography variant="body2" color="text.secondary">AI messages</Typography>
								<Typography variant="h5">{analytics?.ai_messages ?? '-'}</Typography>
							</CardContent>
						</Card>
					</>
				)}
			</Box>

			<Box display="grid" gridTemplateColumns={{ xs: '1fr', md: '1fr 1fr' }} gap={2}>
				<Card>
					<CardHeader title="Recent videos" action={<Button component={Link} href="/videos">All</Button>} />
					<Divider />
					<CardContent>
						{loading ? (
							<Stack spacing={1}>{[...Array(3)].map((_, i) => <Skeleton key={i} height={32} />)}</Stack>
						) : (videos?.length ? (
							<List>
								{videos.map((v: any) => (
									<ListItem key={v.id || v._id} secondaryAction={<Button size="small" component={Link} href={`/videos/${v.id || v._id}`}>Open</Button>}>
										<ListItemText primary={v.title || v.filename || 'Video'} secondary={`Status: ${v.status}`} />
									</ListItem>
								))}
							</List>
						) : (
							<Typography variant="body2" color="text.secondary">No videos yet.</Typography>
						))}
					</CardContent>
				</Card>
				<Card>
					<CardHeader title="Recent sessions" action={<Button component={Link} href="/ai">All</Button>} />
					<Divider />
					<CardContent>
						{loading ? (
							<Stack spacing={1}>{[...Array(3)].map((_, i) => <Skeleton key={i} height={32} />)}</Stack>
						) : (sessions?.length ? (
							<List>
								{sessions.map((s: any) => (
									<ListItem key={s.id || s._id} secondaryAction={<Button size="small" component={Link} href={`/ai/${s.id || s._id}`}>Open</Button>}>
										<ListItemText primary={s.title || 'Session'} secondary={s.created_at || ''} />
									</ListItem>
								))}
							</List>
						) : (
							<Typography variant="body2" color="text.secondary">No sessions yet.</Typography>
						))}
					</CardContent>
				</Card>
			</Box>
		</Stack>
	);
}
