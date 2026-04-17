import { PageSection } from "../components/PageSection"
import secureIcon from "../assets/send-receive-section/icon-secure.svg"
import privateIcon from "../assets/send-receive-section/icon-private.svg"
import sendReceiveIcon from "../assets/send-receive-section/icon-sendreceive.svg"
import handHoldingPhone from "../assets/send-receive-section/hand-hodling-phone.svg"
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

function HandAnimation() {
  const handVariants = {
    hidden: { 
      rotate: 20, 
      y: 80, 
      opacity: 1, 
      scale: 1 
    },
    visible: { 
      rotate: 0, 
      y: 0, 
      opacity: 1, 
      scale: 1,
      transition: { 
        type: "spring", 
        stiffness: 650, 
        damping: 60, 
        mass: 1 
      } 
    }
  }

  const circleVariants = (delay: number) => ({
    hidden: { opacity: 0, scale: 0 },
    visible: { 
      opacity: 1, 
      scale: 1,
      transition: { 
        delay: 0.5 + delay,
        type: "spring", 
        stiffness: 400, 
        damping: 25 
      } 
    }
  })

  return (
    <div className="hand-visual-container">
      <motion.div 
        className="hand-phone-frame"
        variants={handVariants}
        initial="hidden"
        whileInView="visible"
        viewport={{ once: true, margin: "-100px" }}
      >
        <img src={handHoldingPhone} alt="Hand holding phone" className="hand-svg-asset" />
        
        {/* Currency Circles */}
        <motion.div className="currency-circle circle-dollar" variants={circleVariants(0)}>
          <span className="currency-char">$</span>
        </motion.div>
        
        <motion.div className="currency-circle circle-pound" variants={circleVariants(0.1)}>
          <span className="currency-char">£</span>
        </motion.div>
        
        <motion.div className="currency-circle circle-euro" variants={circleVariants(0.2)}>
          <span className="currency-char">€</span>
        </motion.div>
      </motion.div>
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
            <HandAnimation />
          </div>
        </div>
      </div>
    </PageSection>
  )
}

