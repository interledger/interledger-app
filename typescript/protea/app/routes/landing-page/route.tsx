import type { LinksFunction } from "react-router";
import "./colors.css";
import "./tokens/typography.css";
import "./typography.css";
import "./tokens/spacing.css";
import "./layout.css";
import "./tokens/animations.css";
import "./animations.css";
import "./nav.css";
import "./hero.css";
import "./feature.css";
import "./cards.css";
import "./other-features.css";
import "./send-receive.css";
import "./ecosystem.css";
import "./built-with-change.css";
import "./our-ecosystem.css";
import "./footer.css";
import { Layouts } from "~/components/Scaffold";
import { PhoneCarouselProvider } from "./context/PhoneCarouselContext";
import { Nav } from "./components/Nav";
import { HeroSection } from "./components/HeroSection";
import { PhysicalCards } from "./sections/PhysicalCards";
import { OtherFeatures } from "./sections/OtherFeatures";
import { SendReceive } from "./sections/SendReceive";
import { Ecosystem } from "./sections/Ecosystem";
import { BuiltWithChange } from "./sections/BuiltWithChange";
import { OurEcosystem } from "./sections/OurEcosystem";
import { Footer } from "./components/Footer";

export const handle = {
    layout: Layouts.LandingPage,
    scaffold: {
        header: {}
    }
};

export const links: LinksFunction = () => [
    { rel: "preconnect", href: "https://fonts.googleapis.com" },
    { rel: "preconnect", href: "https://fonts.gstatic.com", crossOrigin: "anonymous" },
    { rel: "stylesheet", href: "https://fonts.googleapis.com/css2?family=Poppins:wght@400;600&family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap" }
];

export default function LandingPage() {
    return (
        <PhoneCarouselProvider>
            <div className="landing-shell">
                <Nav />
                <div className="page-content">
                    <HeroSection />
                </div>
                <PhysicalCards />
                <OtherFeatures />
                <SendReceive />
                <Ecosystem />
                <BuiltWithChange />
                <OurEcosystem />
                <Footer />
            </div>
        </PhoneCarouselProvider>
    );
}
