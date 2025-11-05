"use client";

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiUploadVideo, apiGetUserVideos } from '@/lib/api';
import { Alert, Box, Button, Paper, Skeleton, Stack, Typography } from '@mui/material';

export default function VideosPage() {
	const [videos, setVideos] = useState<any[]>([]);
	const [error, setError] = useState<string | null>(null);
	const [loading, setLoading] = useState(false);
	const [initial, setInitial] = useState(true);

	useEffect(() => {
		setInitial(true);
		apiGetUserVideos()
			.then(list => setVideos(Array.isArray(list) ? list : []))
			.catch(e => setError(e?.message || 'Failed to load list'))
			.finally(() => setInitial(false));
	}, []);

	async function onUpload(e: React.ChangeEvent<HTMLInputElement>) {
		const file = e.target.files?.[0];
		if (!file) return;
		setLoading(true);
		setError(null);
		try {
			await apiUploadVideo(file);
			const list = await apiGetUserVideos();
			setVideos(Array.isArray(list) ? list : []);
		} catch (err: any) {
			setError(err?.message || 'Upload failed');
		} finally {
			setLoading(false);
			e.target.value = '';
		}
	}

	return (
		<Stack spacing={3}>
			<Box display="flex" alignItems="center" justifyContent="space-between">
				<Typography variant="h5" fontWeight={600}>Videos</Typography>
				<Button variant="contained" component="label" disabled={loading}>
					{loading ? 'Uploading…' : 'Upload'}
					<input type="file" accept="video/*" hidden onChange={onUpload} />
				</Button>
			</Box>
			{error && <Alert severity="error">{error}</Alert>}
			{initial ? (
				<Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr' }} gap={2}>
					{[...Array(4)].map((_, i) => (
						<Paper key={i} variant="outlined" sx={{ p: 2, borderRadius: 3 }}>
							<Skeleton variant="rounded" height={96} />
						</Paper>
					))}
				</Box>
			) : (videos?.length ? (
				<Box display="grid" gridTemplateColumns={{ xs: '1fr', sm: '1fr 1fr' }} gap={2}>
					{videos.map((v: any) => (
						<Paper key={v.id || v._id} variant="outlined" sx={{ p: 2, borderRadius: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
							<div>
								<Typography fontWeight={600}>{v.title || 'Video'}</Typography>
								<Typography variant="body2" color="text.secondary">Status: {v.status || 'unknown'}</Typography>
							</div>
							<Button component={Link} href={`/videos/${v.id || v._id}`}>Open</Button>
						</Paper>
					))}
				</Box>
			) : (
				<Paper variant="outlined" sx={{ p: 6, textAlign: 'center', borderRadius: 3 }}>
					<Typography color="text.secondary">Your videos will appear here after uploading.</Typography>
				</Paper>
			))}
		</Stack>
	);
}
