import logoImg from '../assets/interledger-wallet-logo.png'

export function Footer() {
  return (
    <footer className='page-footer'>
      <div className='footer-inner'>
        <div className='footer-top'>
          <div className='footer-left'>
            <img
              src={logoImg}
              alt='Interledger Wallet'
              className='footer-logo'
              loading='lazy'
              decoding='async'
            />
            <p className='text-h4'>Do you want to get involved?</p>
            <button className='footer-cta'>
              Contact us
              <svg
                width='20'
                height='20'
                viewBox='0 0 24 24'
                fill='none'
                stroke='currentColor'
                strokeWidth='2.5'
                strokeLinecap='round'
                strokeLinejoin='round'
              >
                <path d='M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6' />
                <polyline points='15 3 21 3 21 9' />
                <line x1='10' y1='14' x2='21' y2='3' />
              </svg>
            </button>
          </div>

          <div className='footer-right'>
            <div className='footer-col'>
              <h4>Useful Links</h4>
              <ul className='footer-links text-body-sm-standard'>
                <li>
                  <a href='#contact'>Contact</a>
                </li>
                <li>
                  <a href='#download'>Download Interledger Wallet</a>
                </li>
                <li>
                  <a href='#testnet'>Test Wallet</a>
                </li>
                <li>
                  <a href='#terms'>Terms & Conditions</a>
                </li>
              </ul>
            </div>

            <div className='footer-col'>
              <h4>Community</h4>
              <ul className='footer-links text-body-sm-standard'>
                <li>
                  <a href='#github'>Contact</a>
                </li>
                <li>
                  <a href='#wm'>Visit WebMonetization</a>
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div className='footer-divider' />

        <div className='footer-bottom'>
          <p>© All Right Reserved by Interledger Foundation</p>
        </div>
      </div>
    </footer>
  )
}
