import { useState, useCallback, useEffect, useRef } from "react"
import { motion, AnimatePresence } from "framer-motion"
import logoImg from "../assets/interledger-wallet-logo.png"
import logoDarkImg from "../assets/interledger-wallet-logo-dark.png"

interface NavProps {
  className?: string
}

/**
 * Fixed navigation bar for the landing page.
 */
export function Nav({ className }: NavProps) {
  const [isOpen, setIsOpen] = useState(false)
  const menuButtonRef = useRef<HTMLButtonElement>(null)
  const toggleMenu = useCallback(() => setIsOpen((prev) => !prev), [])
  const closeMenu = useCallback(() => {
    setIsOpen(false)
    menuButtonRef.current?.focus()
  }, [])

  useEffect(() => {
    if (!isOpen) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") closeMenu()
    }
    document.addEventListener("keydown", handleKeyDown)
    return () => document.removeEventListener("keydown", handleKeyDown)
  }, [isOpen, closeMenu])

  return (
    <>
      <nav className={`nav ${className ?? ""}`}>
        <div className="nav-inner">
          {/* Logo  */}
          <a href="/" className="nav-logo" aria-label="Interledger Wallet home">
            <picture>
              {/* Light theme → use dark logo for contrast */}
              <source srcSet={logoDarkImg} media="(prefers-color-scheme: light)" />
              {/* Dark theme fallback → white logo */}
              <img src={logoImg} alt="" className="nav-logo-icon" />
            </picture>
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
                ref={menuButtonRef}
                type="button"
                className="menu-button"
                onClick={toggleMenu}
                aria-expanded={isOpen}
                aria-controls="mobile-nav"
                aria-label={isOpen ? "Close menu" : "Open menu"}
              >
                <span className="menu-button-icon" aria-hidden="true">
                  {isOpen ? "✕" : "☰"}
                </span>
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Mobile overlay */}
      <AnimatePresence>
        {isOpen && (
          <>
            <div className="nav-backdrop" onClick={closeMenu} aria-hidden="true" />
            <motion.div
              id="mobile-nav"
              className="nav-overlay"
              role="dialog"
              aria-modal="true"
              aria-label="Navigation menu"
              initial={{ opacity: 0, y: -8, scale: 0.98 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, scale: 0.95 }}
              transition={{ duration: 0.2, ease: "easeOut" }}
            >
              <div className="nav-overlay-inner">
                <ul className="nav-overlay-links">
                  <li><a href="#contact" className="nav-link nav-link--mobile" onClick={closeMenu}>Contact</a></li>
                  <li><a href="#login" className="nav-link nav-link--mobile" onClick={closeMenu}>Log in</a></li>
                  <li><a href="#app" className="nav-link nav-link--mobile" onClick={closeMenu}>Use the app</a></li>
                </ul>
                <a href="/signup" className="nav-cta nav-cta--mobile">
                  Get Started <span style={{ marginLeft: "4px" }}>&rarr;</span>
                </a>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  )
}
