"use client";

import { PropsWithChildren, useMemo } from "react";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { createAppTheme } from "@/theme";

export default function ThemeRoot({ children }: PropsWithChildren) {
	const theme = useMemo(() => createAppTheme(), []);
	return (
		<ThemeProvider theme={theme}>
			<CssBaseline />
			{children}
		</ThemeProvider>
	);
}
