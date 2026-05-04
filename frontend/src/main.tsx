import React from "react";
import ReactDOM from "react-dom/client";
import { MantineProvider, createTheme } from "@mantine/core";
import "@mantine/core/styles.css";
import "./styles/app.css";
import App from "./App";

const theme = createTheme({
    primaryColor: "teal",
    fontFamily:
        '"Avenir Next", "SF Pro Display", "Segoe UI", "Helvetica Neue", sans-serif',
    headings: {
        fontFamily:
            '"Avenir Next", "SF Pro Display", "Segoe UI", "Helvetica Neue", sans-serif',
        fontWeight: "800",
    },
    defaultRadius: "lg",
});

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
    <React.StrictMode>
        <MantineProvider theme={theme}>
            <App />
        </MantineProvider>
    </React.StrictMode>,
);
