import { useState, useCallback } from "react"
import { motion, AnimatePresence } from "framer-motion"
import logoImg from "../assets/interledger-wallet-logo.png"

interface NavProps {
  className?: string
}

/**
 * Fixed navigation bar for the landing page.
 * Styled to exactly match the Framer framer reference site.
 */
export function Nav({ className }: NavProps) {
  const [isOpen, setIsOpen] = useState(false)
  const toggleMenu = useCallback(() => setIsOpen((prev) => !prev), [])

  return (
    <>
      <nav className={`nav ${className ?? ""}`}>
        <div className="nav-inner">
          {/* Logo — updated to use PNG icon */}
          <a href="/" className="nav-logo" aria-label="Interledger Wallet home">
            <img src={logoImg} alt="" className="nav-logo-icon" />
          </a>

          {/* Right side container connecting Links and the Button */}
          <div className="nav-right-group">
            {/* Center links — hidden on mobile, visible tablet+ */}
            <ul className="nav-links">
              <li><a href="#contact" className="nav-link">Contact</a></li>
              <li><a href="#login" className="nav-link">Log in</a></li>
              <li><a href="#app" className="nav-link">Use the app</a></li>
            </ul>

            {/* Right actions */}
            <div className="nav-actions">
              <a href="/signup" className="nav-cta">
                Get Started <span style={{ marginLeft: "4px" }}>&rarr;</span>
              </a>
              <button
                type="button"
                className="menu-button"
                onClick={toggleMenu}
                aria-expanded={isOpen}
                aria-label={isOpen ? "Close menu" : "Open menu"}
              >
                <span className="menu-button-icon" aria-hidden="true">
                  {isOpen ? "\u2715" : "\u2630"}
                </span>
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Mobile overlay */}
      <AnimatePresence>
        {isOpen && (
          <motion.div 
            className="nav-overlay" 
            role="dialog" 
            aria-label="Navigation menu"
            initial={{ opacity: 0, y: -8, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.2, ease: "easeOut" }}
          >
            <div className="nav-overlay-inner">
              <ul className="nav-overlay-links">
                <li><a href="#contact" className="nav-link nav-link--mobile" onClick={toggleMenu}>Contact</a></li>
                <li><a href="#login" className="nav-link nav-link--mobile" onClick={toggleMenu}>Log in</a></li>
                <li><a href="#app" className="nav-link nav-link--mobile" onClick={toggleMenu}>Use the app</a></li>
              </ul>
              <a href="/signup" className="nav-cta nav-cta--mobile">
                Get Started <span style={{ marginLeft: "4px" }}>&rarr;</span>
              </a>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  )
}
