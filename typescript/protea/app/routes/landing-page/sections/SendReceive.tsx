import { PageSection } from "../components/PageSection"
import sendMoveyRight from "../assets/send-movey-right.svg"
import secureIcon from "../assets/send-receive-section/icon-secure.svg"
import privateIcon from "../assets/send-receive-section/icon-private.svg"
import sendReceiveIcon from "../assets/send-receive-section/icon-sendreceive.svg"
import { motion } from "framer-motion"

const CheckmarkIcon = () => (
  <div className="checkmark-css" />
)

function StatusItem({ icon, text, index = 0 }: { icon: React.ReactNode, text: string, index?: number }) {
  const animationVariants = {
    hidden: { x: -800, opacity: 1 },
    visible: {
      x: 0,
      opacity: 1,
      transition: {
        type: "spring",
        stiffness: 500,
        damping: 60,
        mass: 1,
        delay: index * 0.1
      }
    }
  }

  return (
    <motion.div
      className="status-item"
      variants={animationVariants}
    >
      <div className="status-icon-container">
        {icon}
      </div>
      <span className="status-text">{text}</span>
      <div className="status-check-container">
        <CheckmarkIcon />
      </div>
    </motion.div>
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
          <motion.div
            className="send-receive-list"
            initial="hidden"
            whileInView="visible"
            viewport={{ once: true, margin: "-50px" }}
          >
            <StatusItem index={0} icon={<img src={secureIcon} />} text="Secure by design" />
            <StatusItem index={1} icon={<img src={sendReceiveIcon} />} text="Send & receive money" />
            <StatusItem index={2} icon={<img src={privateIcon} />} text="Private by default" />
          </motion.div>

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

