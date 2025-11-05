"use client";

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { apiGetVideoStatus, apiGetVideoResult, apiDeleteVideo } from '@/lib/api';
import { Alert, Box, Button, Chip, LinearProgress, Paper, Stack, Typography } from '@mui/material';

export default function VideoDetailPage() {
	const params = useParams<{ id: string }>();
	const router = useRouter();
	const id = params?.id as string;
	const [status, setStatus] = useState<any | null>(null);
	const [result, setResult] = useState<any | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [deleting, setDeleting] = useState(false);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		if (!id) return;
		let mounted = true;
		async function load() {
			try {
				const s = await apiGetVideoStatus(id);
				if (!mounted) return;
				setStatus(s);
				setLoading(false);
				if (s?.status === 'completed') {
					const r = await apiGetVideoResult(id);
					if (!mounted) return;
					setResult(r);
				}
			} catch (e: any) {
				setError(e?.message || 'Failed to load video');
				setLoading(false);
			}
		}
		load();
		const t = setInterval(load, 4000);
		return () => { mounted = false; clearInterval(t); };
	}, [id]);

	async function onDelete() {
		if (!id) return;
		if (!confirm('Delete this video?')) return;
		setDeleting(true);
		try {
			await apiDeleteVideo(id);
			router.replace('/videos');
		} catch (e: any) {
			setError(e?.message || 'Failed to delete');
		} finally {
			setDeleting(false);
		}
	}

	const progress = typeof status?.progress === 'number' ? status.progress : null;
	const statusLabel = status?.status || 'unknown';
	const statusColor: 'default' | 'success' | 'warning' | 'info' =
		statusLabel === 'completed' ? 'success' : statusLabel === 'processing' ? 'warning' : 'info';

	return (
		<Stack spacing={3}>
			<Box display="flex" alignItems="center" justifyContent="space-between">
				<Typography variant="h5" fontWeight={700}>Video</Typography>
				<Button onClick={onDelete} disabled={deleting} variant="outlined" color="error">Delete</Button>
			</Box>
			{error && <Alert severity="error">{error}</Alert>}
			<Paper variant="outlined" sx={{ p: 3, borderRadius: 3 }}>
				<Box display="flex" alignItems="center" gap={2}>
					<Typography fontWeight={600}>Status:</Typography>
					<Chip label={statusLabel} color={statusColor} variant="outlined" sx={{ borderColor: '#00FFC8', color: '#00FFC8' }} />
				</Box>
				{loading && <Box sx={{ mt: 2 }}><LinearProgress /></Box>}
				{progress != null && (
					<Box sx={{ mt: 2 }}>
						<LinearProgress variant="determinate" value={progress} sx={{ height: 8, borderRadius: 2, '& .MuiLinearProgress-bar': { backgroundColor: '#00FFC8' } }} />
						<Typography variant="caption" color="text.secondary" sx={{ ml: 1 }}>{progress}%</Typography>
					</Box>
				)}
			</Paper>
			{result && (
				<Paper variant="outlined" sx={{ p: 3, borderRadius: 3 }}>
					<Typography variant="h6" fontWeight={700}>Result</Typography>
					<Box component="pre" sx={{ mt: 1.5, p: 2, bgcolor: '#0b1220', borderRadius: 2, overflow: 'auto', boxShadow: '0 0 24px rgba(124,58,237,0.12) inset, 0 0 24px rgba(0,255,200,0.08)' }}>
						{JSON.stringify(result, null, 2)}
					</Box>
				</Paper>
			)}
		</Stack>
	);
}
