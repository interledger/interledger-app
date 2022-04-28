import { Container, HomeDecor, Footer, Header } from '~/components'

export default function Page() {
  return (
    <div className='relative w-full overflow-hidden'>
      <Container className='overflow-x-hidden'>
        <Header />
        <main className='flex-grow'>
          <div className='mt-20 mb-12 flex w-[340px] flex-col space-y-8 px-4 sm:mt-44 sm:mb-20 sm:p-8'>
            <span className='font-display text-4xl font-medium leading-normal'>
              Connecting
              <br />
              the Internet
              <br />
              economy
            </span>
          </div>
        </main>
        <Footer />
        <HomeDecor />
      </Container>
    </div>
  )
}
