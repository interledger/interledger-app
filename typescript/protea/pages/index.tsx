import Image from 'next/image'
import styles from '../styles/Home.module.css'

export default function Home() {
  return (
    <div className="w-full h-screen bg-black">
      <div className="w-1/3 md:w-1/4 h-full mx-auto relative">
        <Image src="/fynbos.svg" layout="fill" objectFit="contain" alt="Fynbos"/>
      </div>
    </div>
  )
}
