import { PageSection } from "../components/PageSection"
import sendMoveyRight from "../assets/send-movey-right.svg"
import secureIcon from "../assets/send-receive-section/icon-secure.svg"
import privateIcon from "../assets/send-receive-section/icon-private.svg"
import sendReceiveIcon from "../assets/send-receive-section/icon-sendreceive.svg"

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

export function SendReceive() {
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
          <div className="send-receive-list">
            <StatusItem icon={<img src={secureIcon} />} text="Secure by design" />
            <StatusItem icon={<img src={sendReceiveIcon} />} text="Send & receive money" />
            <StatusItem icon={<img src={privateIcon} />} text="Private by default" />
          </div>

          <div className="send-receive-visual">
            <div className="feature-illustration-container">
              <img
                src={sendMoveyRight}
                alt="Receive Money Illustration"
                className="receive-money-visual-img"
              />
            </div>
          </div>
        </div>
      </div>
    </PageSection>
  )
}
