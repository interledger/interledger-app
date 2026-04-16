import type { LinksFunction } from "react-router";
import colorStylesheet from "./colors.css?url";
import typographyTokens from "./tokens/typography.css?url";
import typographyStylesheet from "./typography.css?url";
import spacingTokens from "./tokens/spacing.css?url";
import layoutStylesheet from "./layout.css?url";
import animationTokens from "./tokens/animations.css?url";
import animationStylesheet from "./animations.css?url";
import navStylesheet from "./nav.css?url";
import heroStylesheet from "./hero.css?url";
import featureStylesheet from "./feature.css?url";
import cardsStylesheet from "./cards.css?url";
import otherFeaturesStylesheet from "./other-features.css?url";
import sendReceiveStylesheet from "./send-receive.css?url";
import { Layouts } from "~/components/Scaffold";
import { PhoneCarouselProvider } from "./context/PhoneCarouselContext";
import { Nav } from "./components/Nav";
import { HeroSection } from "./components/HeroSection";
import { PhysicalCards } from "./sections/PhysicalCards";
import { OtherFeatures } from "./sections/OtherFeatures";
import { SendReceive } from "./sections/SendReceive";

export const handle = {
    layout: Layouts.LandingPage,
    scaffold: {
        header: {}
    }
};

export const links: LinksFunction = () => [
    { rel: "preconnect", href: "https://fonts.googleapis.com" },
    { rel: "preconnect", href: "https://fonts.gstatic.com", crossOrigin: "anonymous" },
    { rel: "stylesheet", href: "https://fonts.googleapis.com/css2?family=Poppins:wght@400;600&family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap" },
    { rel: "stylesheet", href: colorStylesheet },
    { rel: "stylesheet", href: typographyTokens },
    { rel: "stylesheet", href: typographyStylesheet },
    { rel: "stylesheet", href: spacingTokens },
    { rel: "stylesheet", href: layoutStylesheet },
    { rel: "stylesheet", href: animationTokens },
    { rel: "stylesheet", href: animationStylesheet },
    { rel: "stylesheet", href: navStylesheet },
    { rel: "stylesheet", href: heroStylesheet },
    { rel: "stylesheet", href: featureStylesheet },
    { rel: "stylesheet", href: cardsStylesheet },
    { rel: "stylesheet", href: otherFeaturesStylesheet },
    { rel: "stylesheet", href: sendReceiveStylesheet },
];

export default function LandingPage() {
    return (
        <PhoneCarouselProvider>
            <div style={{ backgroundColor: "var(--color-bg-page)", color: "var(--color-text-primary)", minHeight: "100vh" }}>
                <Nav />
                <div className="page-content">
                    <HeroSection />
                </div>
                <PhysicalCards />
                <OtherFeatures />
                <SendReceive />
            </div>
        </PhoneCarouselProvider>
    );
}
