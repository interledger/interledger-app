import { Container, Decor, Footer, Header } from 'components'
import { NextPage } from 'next'

const HomePage: NextPage = () => {
  return (
    <div className='relative overflow-hidden w-full'>
      <Container className=' overflow-x-hidden'>
        <Header />
        <main className='flex-grow'>
          <div className='flex flex-col px-4 sm:p-8 mt-20 sm:mt-44 mb-12 sm:mb-20 space-y-8 leading-normal w-[340px]'>
            <span className='font-display text-[40px] leading-normal font-medium'>
              Connecting
              <br />
              the Internet
              <br />
              economy
            </span>
            <span className='text-2xl leading-normal'>Coming soon.</span>
          </div>
        </main>
        <Footer />
        <Decor />
      </Container>
    </div>
  )
}

export default HomePage
