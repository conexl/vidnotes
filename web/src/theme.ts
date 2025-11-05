import { createTheme } from '@mui/material/styles';

export const createAppTheme = () =>
	createTheme({
		palette: {
			mode: 'dark',
			primary: { main: '#00FFC8' },
			secondary: { main: '#7C3AED' },
			background: {
				default: '#0A0B0F',
				paper: '#0F1016',
			},
			text: {
				primary: '#E8EAED',
				secondary: '#9AA0A6',
			},
		},
		shape: { borderRadius: 14 },
		components: {
			MuiPaper: { styleOverrides: { root: { borderRadius: 14 } } },
			MuiButton: { styleOverrides: { root: { borderRadius: 12 } }, defaultProps: { disableElevation: true } },
			MuiCard: { styleOverrides: { root: { borderRadius: 16 } } },
		},
	});
