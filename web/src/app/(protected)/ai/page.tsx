"use client";

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiGetSessions, apiCreateSession, apiGetUserVideos } from '@/lib/api';
import { Alert, Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, MenuItem, Paper, Select, Skeleton, Stack, TextField, Typography } from '@mui/material';

export default function AISessionsPage() {
	const [sessions, setSessions] = useState<any[] | null>([]);
	const [title, setTitle] = useState('');
	const [creating, setCreating] = useState(false);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	const [videos, setVideos] = useState<any[]>([]);
	const [videoId, setVideoId] = useState<string>('');
	const [dialogOpen, setDialogOpen] = useState(false);

	async function reload() {
		setLoading(true);
		try {
			const [list, vids] = await Promise.all([apiGetSessions(), apiGetUserVideos()]);
			setSessions(Array.isArray(list) ? list : []);
			setVideos(Array.isArray(vids) ? vids : []);
		} catch (e: any) {
			setError(e?.message || 'Failed to load');
			setSessions([]);
		} finally {
			setLoading(false);
		}
	}

	useEffect(() => {
		reload();
	}, []);

	function openCreateDialog() { setDialogOpen(true); }
	function closeCreateDialog() { setDialogOpen(false); setTitle(''); setVideoId(''); }

	async function onCreate(e: React.FormEvent) {
		e.preventDefault();
		if (!videoId) { setError('Select a video'); return; }
		setCreating(true);
		setError(null);
		try {
			await apiCreateSession({ title: title || undefined, video_id: videoId });
			closeCreateDialog();
			await reload();
		} catch (e: any) {
			setError(e?.message || 'Failed to create session');
		} finally {
			setCreating(false);
		}
	}

	return (
		<Stack spacing={3}>
			<Box display="flex" alignItems="center" justifyContent="space-between">
				<Typography variant="h5" fontWeight={600}>AI Sessions</Typography>
				<Button variant="contained" onClick={openCreateDialog}>Create</Button>
			</Box>
			{error && <Alert severity="error">{error}</Alert>}
			{loading ? (
				<Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr' }} gap={2}>
					{[...Array(4)].map((_, i) => (
						<Paper key={i} variant="outlined" sx={{ p: 2, borderRadius: 3 }}>
							<Skeleton variant="rounded" height={96} />
						</Paper>
					))}
				</Box>
			) : (sessions && sessions.length > 0 ? (
				<Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr' }} gap={2}>
					{(sessions ?? []).map((s: any) => (
						<Paper key={s.id || s._id} variant="outlined" sx={{ p: 2, borderRadius: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
							<div>
								<Typography fontWeight={600}>{s.title || 'Session'}</Typography>
								<Typography variant="body2" color="text.secondary">{s.created_at || ''}</Typography>
							</div>
							<Button component={Link} href={`/ai/${s.id || s._id}`}>Open</Button>
						</Paper>
					))}
				</Box>
			) : (
				<Paper variant="outlined" sx={{ p: 6, textAlign: 'center', borderRadius: 3 }}>
					<Typography color="text.secondary">No sessions yet. Create one.</Typography>
				</Paper>
			))}

			<Dialog open={dialogOpen} onClose={closeCreateDialog} fullWidth maxWidth="sm">
				<DialogTitle>New AI session</DialogTitle>
				<DialogContent>
					<Stack component="form" spacing={2} sx={{ mt: 1 }} onSubmit={onCreate}>
						<TextField label="Title (optional)" value={title} onChange={e => setTitle(e.target.value)} fullWidth />
						<Select displayEmpty value={videoId} onChange={e => setVideoId(String(e.target.value))} fullWidth>
							<MenuItem value=""><em>Select a video</em></MenuItem>
							{videos.map(v => (
								<MenuItem key={v.id || v._id} value={v.id || v._id}>{v.title || v.filename || 'Video'}</MenuItem>
							))}
						</Select>
						<Button type="submit" variant="contained" disabled={creating || !videoId}>Create</Button>
					</Stack>
				</DialogContent>
				<DialogActions>
					<Button onClick={closeCreateDialog}>Cancel</Button>
				</DialogActions>
			</Dialog>
		</Stack>
	);
}
