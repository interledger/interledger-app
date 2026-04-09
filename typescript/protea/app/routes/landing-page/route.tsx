import type { LinksFunction } from "react-router";
import colorStylesheet from "./colors.css?url";
import typographyTokens from "./tokens/typography.css?url";
import typographyStylesheet from "./typography.css?url";
import spacingTokens from "./tokens/spacing.css?url";
import layoutStylesheet from "./layout.css?url";
import { Layouts } from "~/components/Scaffold";

export const handle = {
    layout: Layouts.LandingPage,
    scaffold: {
        header: {}
    }
};

export const links: LinksFunction = () => [
    { rel: "preconnect", href: "https://fonts.googleapis.com" },
    { rel: "preconnect", href: "https://fonts.gstatic.com", crossOrigin: "anonymous" },
    { rel: "stylesheet", href: "https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap" },
    { rel: "stylesheet", href: colorStylesheet },
    { rel: "stylesheet", href: typographyTokens },
    { rel: "stylesheet", href: typographyStylesheet },
    { rel: "stylesheet", href: spacingTokens },
    { rel: "stylesheet", href: layoutStylesheet },
];

export default function LandingPage() {
    return (
        <div style={{ backgroundColor: "var(--color-bg-page)", color: "var(--color-text-primary)", minHeight: "100vh" }}>
            <h1 className="text-h1">Landing Page Route</h1>
            <p className="text-body-lg-standard">Color and typography systems loaded with placeholder values.</p>
        </div>
    );
}
