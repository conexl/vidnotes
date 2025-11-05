"use client";

import { useEffect, useRef, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { apiGetSession, apiSendMessage, apiDeleteSession } from '@/lib/api';
import { Alert, Box, Button, IconButton, Paper, Stack, TextField, Tooltip, Typography } from '@mui/material';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';

export default function AISessionChatPage() {
	const params = useParams<{ id: string }>();
	const router = useRouter();
	const id = params?.id as string;
	const [session, setSession] = useState<any | null>(null);
	const [messages, setMessages] = useState<any[]>([]);
	const [input, setInput] = useState('');
	const [sending, setSending] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [deleting, setDeleting] = useState(false);
	const [summaryExpanded, setSummaryExpanded] = useState(false);
	const endRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		if (!id) return;
		apiGetSession(id)
			.then((s: any) => {
				setSession(s);
				if (s?.messages) setMessages(s.messages);
			})
			.catch((e: any) => setError(e?.message || 'Failed to load session'));
	}, [id]);

	useEffect(() => {
		endRef.current?.scrollIntoView({ behavior: 'smooth' });
	}, [messages]);

	function isLong(text: string): boolean {
		const lines = text.split(/\r?\n/);
		return lines.length > 3 || text.length > 320;
	}

	async function onSend(e: React.FormEvent) {
		e.preventDefault();
		if (!input.trim()) return;
		setSending(true);
		setError(null);
		const userMsg = { role: 'user', content: input };
		setMessages(prev => [...prev, userMsg]);
		setInput('');
		try {
			const res: any = await apiSendMessage(id, { content: userMsg.content });
			if (res?.message) {
				const assistantMsg = { role: 'assistant', content: res.message };
				setMessages(prev => [...prev, assistantMsg]);
			}
		} catch (e: any) {
			setError(e?.message || 'Failed to send message');
		} finally {
			setSending(false);
		}
	}

	async function onDelete() {
		if (!id) return;
		if (!confirm('Delete this session?')) return;
		setDeleting(true);
		try {
			await apiDeleteSession(id);
			router.replace('/ai');
		} catch (e: any) {
			setError(e?.message || 'Failed to delete');
		} finally {
			setDeleting(false);
		}
	}

	const summary = session?.summary || '';
	const summaryLong = summary ? isLong(summary) : false;
	const summaryClampedStyle = !summaryExpanded && summaryLong
		? { display: '-webkit-box', WebkitLineClamp: 3 as any, WebkitBoxOrient: 'vertical' as any, overflow: 'hidden' }
		: undefined;

	return (
		<Stack spacing={2} sx={{ height: 'calc(100dvh - 120px)' }}>
			<Box display="flex" alignItems="center" justifyContent="space-between">
				<div>
					<Typography variant="h5" fontWeight={700}>{session?.title || 'Session'}</Typography>
					{summary && (
						<>
							<Typography variant="body2" color="text.secondary" sx={summaryClampedStyle}>{summary}</Typography>
							{summaryLong && (
								<Box sx={{ mt: 0.5 }}>
									<Button size="small" onClick={() => setSummaryExpanded(v => !v)} sx={{ color: '#00FFC8', px: 0 }}>
										{summaryExpanded ? 'Show less' : 'Show more'}
									</Button>
								</Box>
							)}
						</>
					)}
				</div>
				<Tooltip title="Delete session">
					<span>
						<IconButton onClick={onDelete} disabled={deleting} color="error">
							<DeleteOutlineIcon />
						</IconButton>
					</span>
				</Tooltip>
			</Box>
			<Paper variant="outlined" sx={{ p: 2, borderRadius: 3, flex: 1, overflowY: 'auto', position: 'relative', background: (t) => t.palette.background.paper }}>
				{messages.map((m, i) => (
					<Box key={i} sx={{ display: 'flex', justifyContent: m.role === 'user' ? 'flex-end' : 'flex-start', mb: 1.25 }}>
						<Box
							sx={{
								maxWidth: '80%',
								px: 2,
								py: 1.25,
								borderRadius: 3,
								color: m.role === 'user' ? 'common.white' : 'text.primary',
								bgcolor: m.role === 'user' ? '#111' : 'action.hover',
								boxShadow: m.role === 'user' ? '0 0 12px rgba(0,255,200,0.35), 0 0 2px rgba(0,255,200,0.8) inset' : '0 0 0 rgba(0,0,0,0)',
								position: 'relative',
								'&::after': m.role === 'assistant' ? {
									content: '""',
									position: 'absolute',
									top: 0,
									left: 0,
									right: 0,
									bottom: 0,
									pointerEvents: 'none',
									background: 'repeating-linear-gradient(90deg, transparent 0 2px, rgba(0,255,200,0.08) 2px 3px)',
									mixBlendMode: 'overlay',
								} : undefined,
							}}
						>
							<Typography sx={{ whiteSpace: 'pre-wrap' }}>{m.content}</Typography>
						</Box>
					</Box>
				))}
				<div ref={endRef} />
			</Paper>
			{error && <Alert severity="error">{error}</Alert>}
			<Box component="form" onSubmit={onSend} display="flex" gap={1}>
				<TextField
					value={input}
					onChange={(e) => setInput(e.target.value)}
					placeholder="Write a message…"
					fullWidth
					variant="outlined"
					InputProps={{ sx: { borderRadius: 3 } }}
				/>
				<Button type="submit" variant="contained" disabled={sending} sx={{ borderRadius: 3 }}>Send</Button>
			</Box>
		</Stack>
	);
}
