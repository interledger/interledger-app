import { useRef } from "react"
import { useInView } from "framer-motion"
import { PageSection } from "../components/PageSection"
import secureIcon from "../assets/send-receive-section/icon-secure.svg"
import privateIcon from "../assets/send-receive-section/icon-private.svg"
import sendReceiveIcon from "../assets/send-receive-section/icon-sendreceive.svg"
import handHoldingPhone from "../assets/send-receive-section/hand-hodling-phone.svg"

const CheckmarkIcon = () => (
  <div className="checkmark-css" />
)

function StatusItem({ icon, text }: { icon: React.ReactNode, text: string }) {
  return (
    <div className="status-item">
      <div className="status-icon-container">
        {icon}
      </div>
      <span className="status-text">{text}</span>
      <div className="status-check-container">
        <CheckmarkIcon />
      </div>
    </div>
  )
}

function HandAnimation() {
  const ref = useRef<HTMLDivElement>(null)
  const isInView = useInView(ref, { once: true, margin: "-100px" })

  return (
    <div ref={ref} className={`hand-visual-container ${isInView ? "is-visible" : ""}`}>
      <div className="hand-phone-frame">
        <img src={handHoldingPhone} alt="Hand holding phone" className="hand-svg-asset" />
        
        {/* Currency Circles */}
        <div className="currency-circle circle-dollar">
          <span className="currency-char">$</span>
        </div>
        
        <div className="currency-circle circle-pound">
          <span className="currency-char">£</span>
        </div>
        
        <div className="currency-circle circle-euro">
          <span className="currency-char">€</span>
        </div>
      </div>
    </div>
  )
}

export function SendReceive() {
  const ref = useRef<HTMLDivElement>(null)
  const isInView = useInView(ref, { once: true, margin: "-50px" })
  return (
    <PageSection className="send-receive-section">
      <div className="send-receive-container">
        <header className="send-receive-header">
          <h2 className="send-receive-title">Today, the wallet lets you send and receive money</h2>
          <p className="send-receive-description">
            Starting with the fundamentals keeps the wallet reliable, while allowing it to scale responsibly.
            Our aim is to remain true to the protocol it supports over the long run.
          </p>
        </header>

        <div className="send-receive-content">
          <div
            ref={ref}
            className={`send-receive-list ${isInView ? "is-visible" : ""}`}
          >
            <StatusItem icon={<img src={secureIcon} />} text="Secure by design" />
            <StatusItem icon={<img src={sendReceiveIcon} />} text="Send & receive money" />
            <StatusItem icon={<img src={privateIcon} />} text="Private by default" />
          </div>

          <div className="send-receive-visual">
            <HandAnimation />
          </div>
        </div>
      </div>
    </PageSection>
  )
}

