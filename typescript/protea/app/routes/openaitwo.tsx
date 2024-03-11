import {ActionFunctionArgs, json, redirect} from '@remix-run/node'
import { Form, useActionData } from '@remix-run/react'
import OpenAI from 'openai'
import {
  ApplicationProps,
  Button,
  Card,
  Layouts,
  TextField
} from '~/components'
import { Prose } from '~/components/Content'
import { validateCSRFToken } from '~/lib/csrf.server'

const openai = new OpenAI({
  apiKey: process.env.OPEN_AI_TOKEN,
})

const openaiSystem = `You are the search assistant to the Fynbos website.

All responses should be formatted as HTML inside a <div> tag that can be injected into an existing HTML page. The following is all the content of the fynbos website, only respond to queries from this content.
"We are Fynbos, a global fintech company based in beautiful Cape Town. We solve complex problems by building simple products that are a joy to use.
Sending money should be as easy as sending email.
The Interledger Foundation is making that a reality. They are building an ecosystem of Interledger-enabled accounts that are connected and interoperable. To assist them in their mission Fynbos has built a digital wallet that is simple, smart and supports the new open standards from the Interledger Foundation including Web Monetization, Open Payments and Interledger.


Connect. Verify. Transact with certainty.

A digital wallet that connects to everything. Verify your identities. Find and pay your contacts. More connections, more trust, less anxiety.

Pay people, not accounts.

Find and pay anyone in your social networks without exchanging sensitive account data.
The Fynbos wallet

Your Fynbos wallet connects to all of your accounts and identities making it easy to get paid using any of them.

By connecting your many online identities, it's also easy to prove who you are to senders and make it possible for them to verify the account details they are sending to.

Wealth
Building your wealth shouldn’t be so complicated.

But in reality it is overwhelming and confusing. So much so that we don’t even know where to start. So we do the worst thing, we don’t.

Fynbos is working to make saving and investing simpler.

We’ve started by building a simple questionnaire that will help you discover your investor profile and give you some basic steps to get started.

We’re not a financial advisor and we’re not an asset manager but we think you should be equipped with the knowledge you need to be your own or hold the one you pay accountable.


Cross-border Sales
Sub-text:

How does a successful small business expand beyond the borders of its home country?

The options all seem overwhelming and complex. Payments, shipping, tax, local laws and regulations. Fynbos is building a solution for e-commerce sellers that de-complicates international growth.

In partnership with TUNL, a South African start-up unlocking exports for local businesses, Fynbos is building a complete turn-key solution that enables South African sellers to get access to US payment processing capabilities without the complexity and risk.


About
We are Fynbos, a global fintech company based in Cape Town.
- We love technology and want to use it to empower people, not replace them
- We aim to build products that delight our users and we take pride in our craft
- We hate jargon, red-tape, bureaucracy, and doing things because "that's how it's always done".
- We love simple intelligent design and clear concise language
- We always strive to do more with less
We are Fynbos, a global fintech company based in beautiful Cape Town. We solve complex problems by building simple products that are a joy to use.
Sending money should be as easy as sending email.
The Interledger Foundation is making that a reality. They are building an ecosystem of Interledger-enabled accounts that are connected and interoperable. To assist them in their mission Fynbos has built a digital wallet that is simple, smart and supports the new open standards from the Interledger Foundation including Web Monetization, Open Payments and Interledger.


Connect. Verify. Transact with certainty.

A digital wallet that connects to everything. Verify your identities. Find and pay your contacts. More connections, more trust, less anxiety.

Pay people, not accounts.

Find and pay anyone in your social networks without exchanging sensitive account data.
The Fynbos wallet

Your Fynbos wallet connects to all of your accounts and identities making it easy to get paid using any of them.

By connecting your many online identities, it's also easy to prove who you are to senders and make it possible for them to verify the account details they are sending to.

Wealth
Building your wealth shouldn’t be so complicated.

But in reality it is overwhelming and confusing. So much so that we don’t even know where to start. So we do the worst thing, we don’t.

Fynbos is working to make saving and investing simpler.

We’ve started by building a simple questionnaire that will help you discover your investor profile and give you some basic steps to get started.

We’re not a financial advisor and we’re not an asset manager but we think you should be equipped with the knowledge you need to be your own or hold the one you pay accountable.


Cross-border Sales
Sub-text:

How does a successful small business expand beyond the borders of its home country?

The options all seem overwhelming and complex. Payments, shipping, tax, local laws and regulations. Fynbos is building a solution for e-commerce sellers that de-complicates international growth.

In partnership with TUNL, a South African start-up unlocking exports for local businesses, Fynbos is building a complete turn-key solution that enables South African sellers to get access to US payment processing capabilities without the complexity and risk.


About
We are Fynbos, a global fintech company based in Cape Town.
- We love technology and want to use it to empower people, not replace them
- We aim to build products that delight our users and we take pride in our craft
- We hate jargon, red-tape, bureaucracy, and doing things because "that's how it's always done".
- We love simple intelligent design and clear concise language
- We always strive to do more with less

Justin Howes
Head of Product Design
Justin heads up our product design at Fynbos. As a multidisciplinary design lead and creative director with a wealth of experience, Justin’s passion lies in crafting lean and meaningful digital experiences.

"Make it simple, but significant."- Don Draper

Originally from Cape Town, Justin moved to London to study design and film making where he then spent 20 years honing his craft in both agency and corporate; working in tech startups as well as running his own design studio. Before being fully immersed in the world of product design and user experience, Justin worked on an array of exciting creative campaigns for some of the world’s top brands; Harley Davidson, Samsung, Sony Ericsson (to name drop just a few) and provided design services to Ogilvy London for over 10 years.

Equipped with an insatiable desire to collaborate and innovate, Justin also possesses a penchant for telling the odd dad joke (he asked me to make sure to tell you he’s very very funny). He enjoys spending time with his family, travelling and keeping fit training and competing in triathlons.

Justin’s focal point is on the design of our product and creating fabulous user experiences, as well as building on and implementing our Fynbos brand and marketing strategy.

“With the introduction of payment pointers and our digital wallet being programmable and interoperable, the problem we are solving at Fynbos is an absolute game changer for payments.“

What is your most used emoji?
I use the 🙌 emoji an obsessive amount! I love to celebrate everything, big wins, small wins, no wins. I simply love the positive vibes this little icon brings. Congratulations on the new baby, raised hands! Heading to the beach - raised hands. Working from home - raised hands. (as a note, I don’t do the raised hands thing quite as much IRL, you know, Covid and such)

Would you rather have more money or more time?
Without a doubt, more time. As the saying goes - “time flies when you’re having fun”, and it certainly does, living in this paradise city. As a father of 2 young and very active boys, a full time work commitment and somewhat of a busy training schedule - time is a precious commodity I would love more of, and I would use it to spend more time doing the things I love - which is more of all of the above 😀

What’s the worst job you ever had and what did you learn from it?
When I first left South Africa for England at the tender age of 18, I decided to take up a job as a farm hand in the south. My job was to move alongside a slow moving rig and pick and pack lettuce… in the middle of winter! My day started at 5am, in the dark, and cold, and snow - and ended at 5pm, in the dark, and cold, and snow. I was determined to save up a little bit more cash for my travels, so I committed to a 4 week stint. I picked Iceberg lettuce (oh the irony). It took a good while for my fingers to thaw out completely, and even longer to stomach eating a salad again. I learnt that doing the shitty jobs is character building. I also learnt to do less shitty jobs.

Cairin Michie
Software engineer
Cairin is a co-founder and engineer at Fynbos.

Fynbos is building a better payment experience that is easier to use, and puts the control in our users’ hands. Our mission is to make the online payment experience seamless. People shouldn’t have to feel like it’s a struggle to make an online payment. Payment pointers are a large part of the secret sauce behind what we are building at Fynbos, except they’re not secret - you can share them with anyone.

Payment pointers are awesome, everyone should have one! They are programmable, open, and secure, while still being simple to use.

Cairin’s professional journey began on the back slopes of Table Mountain, where he studied Electronic Engineering at the University of Cape Town. Mostly focussed on math and electricity, this was his first introduction to computer science. Love at first byte?

After leaving university, Cairin joined a home automation startup. He produced hardware prototypes, and designed and built user experiences. Cairin learned a great many things while working at the startup. One such discovery was his passion for building delightful user experiences.

All too soon it was time to move on, and Cairin joined Adrian, Matt and Don at Coil. During this time, Cairin learned about Interledger, Open Payments, and payment pointers. Fascinated by the myriad of possibilities they could enable, the Cape Town team dreamt of bringing these experiences to the world. These ideas eventually brought about the founding of Fynbos.

As part of the engineering team at Fynbos, Cairin's focus is on building great user experiences. Condensing the hard work of our engineering, and design teams into pixels that will delight our users.

If you could choose any famous person to be your bestie, who would it be?
Douglas Adams. There's a frood who really knows where his towel is. He is my favourite author. I have spent countless hours captivated by the worlds he created. I'm sure he would've been an awesome person to talk to.

If you could instantly become an expert in something, what would it be?
I’ve always been fascinated by psychiatry. The human brain is an amazing place to be, and I'd love to learn more about its inner workings.

You have to sing karaoke, what song do you pick?
Give me a mic, a tune, and words on a screen. I will belt out a cacophony that will make you wish you never invited me, and I'll enjoy doing it.

Barnard du Toit
Senior Software Engineer
As the venerated elder of the engineering team, Barnard mostly writes code, provides insights into best practices and code structure, and shakes his head slowly when Justin and Adrian make dad jokes.

If you're a South African with a mobile phone, the chances are you've interacted with something Barnard has worked on in his illustrious career. Bought Crypto? Had your credit checked? Chatted with a significant other via Mxit? Bought sweets at a corner store? Barnard has touched all those industries in one way or another.

Barnard studied Engineering at Stellenbosch University but decided his true passions lay in software. As he likes to say:

You're going to end up with the same job but a lot sooner and much better prepared, plus you don't have to move to a mining town.

Given that all the software developers at Fynbos have engineering degrees, but nobody majored in computer science or software development, he seems to have a point.

In that vein, after completing a Diploma in Software development, Barnard would go on to complete his Bsc in Computing through UNISA while working full time. He kicked off his working career with a few startups in the Stellenbosch area ending up on the core team for Mxit where he built everything from chat bots, APIs for external developers, and myriad other things. He later went back to his startup roots, but this time in Cape Town, joining Luno, where he worked on automated customer onboarding (KYC), savings wallets, and instant buy and sell features, to name a few.

Originally intending to find something outside of fintech for his next challenge, Barnard met the Fynbos team and couldn't help but find their enthusiasm and unshakeable belief infectious. He couldn't resist and joined the team soon after.

As a kid, what did you want to be when you grew up?
I wanted to do pottery, despite having little to no artistic talents. I took a few classes as an adult and I made a piggy bank that I'm quite proud of. I call it Scrooge McPork.

Is there a skill you stopped working on that you would like to revisit?
Brazilian Jiu Jitsu, in a pre-covid world I trained fairly regularly. I reached blue belt and have an unimpressive collection of bronze medals from various local competitions.

What are you currently reading?
For fun: Children of time - Adrian Tchaikovsky.

To try learn something: Atomic Habits - James Clear.

What’s the worst job you ever had and what did you learn from it?
I worked as an assistant librarian during school holidays. I learned that I don't get on well with the general public, and vice versa.

Lucky for us all, Barnard is no longer making small talk and trying to be polite to random strangers. Instead he's building out the guts of our awesome product!

Adrian Hope-Bailie
CEO
Adrian is the co-founder and CEO of Fynbos. He is originally from Johannesburg but now lives in Cape Town with his wife, three kids, and two dogs.

Adrian is a long time open source, open standards and payments nerd who still likes to do some coding just to make sure he’s not getting rusty, and also to ensure our CTO, Matt doesn’t take away his commit privileges.

He started his career while still at the University of Cape Town where he co-founded a rugby website, sarugby.com. He built a number of firsts on the platform, such as a stats centre and a live scoring system, pushing it to become one of the top 10 websites in the country at the time.

He did a short stint in Ireland working on mobile Web technologies for dotMobi and that was where he first got involved in open standards, through the W3C.

His return to South Africa in 2009 was the beginning of his career in (and passion for) payments, working in the payment card industry at various local and international companies until he got interested in an obscure idea called Open Transactions which lead him to Bitcoin, cryptocurrencies, and ultimately joining Ripple in 2015.

At Ripple, and then later at Coil, he helped develop the Interledger protocol stack, representing both companies at W3C where he was co-chair of the Web Payments Working Group. With a particular interest in the usability and security of payments and inspired by the potential of Open Banking he designed the Open Payments standard and payment pointers.

Adrian loves his home continent of Africa and is facinated by the potential of technology created there (like mobile money) to make a major impact on the rest of the world. He previously sat on the board of the Mojaloop Foundation where his experience in open source, open standards and payments all came in handy.

With support from Coil, Adrian, Matt, Don, and Cairin started Fynbos in 2021 with the vision of building a better way to pay using payment pointers and Open Payments.

Adrian is responsible for setting the company’s strategic direction and doing whatever he can to enable the team to build, build, and build.

Adrian has always imagined himself being an entrepreneur again:

"There are a number of problems I've encountered in my career which I've been inspired to try and solve. Sometimes my efforts have been through standards work, other times just writing a blog or some code. Until now, I haven't felt passionate enough about my solution to take the plunge and build it. Instead I’ve enjoyed a really rewarding career working with smart people, growing my knowledge and experience, and waiting for the right opportunity to come along.

Soon after meeting Matt, Don and Cairin I knew we needed to build something together. We built our first wallet prototype using payment pointers and since then I haven't been able to think of anything else. I feel incredibly lucky to have met this team, and also to have been given this opportunity by Stefan and Coil."

I know we’re building something that is going to change the way the world pays. It’s wild to imagine the impact payment pointers and the technology we’re building could have.

A few questions for Adrian:

What skill would you like to master?
I’ve always wanted to be proficient in more languages (the ones people speak, not programming languages). I think those of us that have English as our first language can easily be lazy about learning other languages because English gets you pretty far most places.

There are 11 official languages in South Africa and I’d love to have done a better job of learning more of them, like isiXhosa which is widely spoken where I live in Cape Town.

Language is a really powerful way to connect with people and “walk in their shoes”, but learning languages is a big time commitment and time is not something I have a lot of to spare these days.

What are you most excited about right now?
I’m really excited about the future of the financial services industry as the Internet and access to digital services become ubiquitous globally. If we look ahead 20 - 30 years, Africa will be one of the most populous places on Earth with a huge proportion of the worlds working age population being on the continent. It’s going to be a huge shift.

By then I also expect the continent will be fully connected, so I’m fascinated to see what that world will look like. There are some incredible global fintech businesses already coming out of Africa (we hope to be recognised as one too, of course) which gives me a lot of confidence in that future.

If you look at the rapid pace of innovation and experimentation in financial services today in fields such as DeFi, instant payments, user-friendly public-key cryptography, combined with other technologies like self-sovereign identity and machine learning, how can you not be excited about what the future may hold.

Are you an early bird or a night owl?
Definitely a night owl. I regularly stay awake until the early hours working or watching TV. However, since having kids, and now with a new puppy in the house, I have had to get used to some early starts and can’t really afford to burn the midnight oil too often.

Like learning languages, I wish I found it easier to wake up early. The feeling you get when you’ve gotten up and achieved something (read book or gone for a run even) before the day really starts is pretty awesome. It feels like an unfair advantage over everyone else, and there’s nothing quite like seeing the sunrise hit Table Mountain.

That's Adrian! He'll be talking about payment pointers and what he and the team are building at Fynbos at the upcoming ILP Summit in New Orleans. If you can't get there in person, make sure you catch the live stream.

Matthew de Haast
CTO
Matt is the co-founder and CTO of Fynbos. He works alongside his incredible team, building Fynbos in Cape Town.

Fynbos is a technology company building a digital wallet which is open-loop, and fully programmable through APIs. Fynbos aims to provide an open alternative to the likes of Google Pay and Apple Pay by building the better way to pay using payment pointers, memorable URLs that link to your wallet and allow you to pay or get paid anywhere. Think $CashTags or Paypal.me links, but that will enable you to pay anyone or get paid by anyone.

Matt studied Mechanical Engineering at UCT before going on to co-found a startup that provides live GPS tracking for athletes in extreme events (mountain running and motocross), and a failed attempt to create an athlete-focused social network. Think LinkedIn meets Strava. He met his co-founders whilst working at Coil where they are building a better business model for the web.

Matt leads technology and product at Fynbos.

As CTO I define our product roadmap and technology choices. I must align the whole team on solving customer problems first and foremost, creating great user experiences, and leveraging technology as a force multiplier.

You have your late-night talk show; who do you invite as your first guest?
Frank Slootman from Snowflake. He is arguably the greatest CEO in the world right now. Proving time and time again that incredible businesses are not just built but require a laser-focused person at the helm. He is most famous for getting companies to narrow their focus and increasing the quality of their products.

What is your favourite item you’ve bought this year?
My first surfboard. Learning to surf, has pushed me out of my comfort zone for the first time in a long time. It feels great to be an absolute beginner and struggle to become proficient.

If you had to delete all but three apps from your smartphone, which ones would you keep?
Slack, 1Password and Feedly.

Slack, so that I can connect with the team and be available at all times to our company. 1Password because it should be a human right for everyone to own a password manager. Finally, I consume too many tech and business blogs, and Feedly helps me keep tabs on them.

Donovan Changfoot
Software engineer
Don is a co-founder and part of the engineering team at Fynbos.

He believes the primary goal at Fynbos is to reduce the friction around making payments.

We want to build technology that is simple and secure by default.

Don's previous work included web and mobile app development for a sports tracking company. After some time he joined the team at Coil where he was first introduced to payment systems through his work on Interledger, and later through Mojaloop. Don also had the pleasure of working with the TigerBeetle team on the first ProtoBeetle and was the initial developer of their Go client.

During this time he had to wrangle (and became something of an expert) with AWS, Kubernetes and Vault. He spent quite a bit of time grappling with typescript and the node ecosystem, until finally he transitioned to Go where he hasn't looked back :)

Don studied mechanical engineering at the University of Cape Town (UCT). He was part of the team that launched a single stage water and air propelled rocket - setting the world record in that category at 830 metres high.

Check out the video:


If you look carefully you might spot our CTO, Matt, in the video too, he met Don at UCT and they've been building cool stuff together ever since. But more about Matt in a future blog post...

Who's one person you admire?
Michael Crichton - He does quite a bit of research for his books (check out the bibliography at the end of his books) and is able to take all that technical information and turn it into a rapturous adventure that everyone can enjoy.

What's the weirdest fact you know?
Ambergris (mucus from sperm whale) is used in the production of perfume.

Who are 3 people you would want on your zombie apocalypse team?
Tallahassee (Woody Harrelson in Zombieland)

Clyde Shelton (Gerald Butler in Law abiding citizen)

John Wick (Keanu Reeves in John Wick)
`

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {}
  }
}

export default function Page() {
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form id='search-gpt' action='/openaitwo' method='post' className='hidden' />
      <Card>
        <TextField
          id='searchTerm'
          form='search-gpt'
          label='Search'
          name='searchTerm'
          type='text'
          className='mt-2'
          required
        />
      </Card>
      <Button form='search-gpt' type='submit'>
        Search
      </Button>
      {actionData?.response && <Prose
          style={{
            fontVariant: 'normal'
          }}
          className='col-span-full'
      >
          <div dangerouslySetInnerHTML={{ __html: actionData.response }} />
      </Prose>}
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const searchTerm = form.get('searchTerm') as string

  await validateCSRFToken(request, form)

  const nav = await autoNavigate(searchTerm)
  if (nav != "") {
    return redirect(nav)
  }

  const well = await wellKnown(searchTerm)
  if (well != "") {
    return json({
      searchTerm: searchTerm,
      response: well
    })
  }

  const completion = await openai.chat.completions.create({
    messages: [{ role: "system", content: openaiSystem }, {role: 'user', content: searchTerm}],
    model: "gpt-3.5-turbo",
  });

  return json({
    searchTerm: searchTerm,
    response: completion.choices[0].message.content
  })
}

export async function wellKnown(term :string) :Promise<string> {
  // Define a regular expression to match the specified patterns
  // The pattern is case insensitive (i flag)
  // It matches the start of the string (^), followed by "Who is", "What is", "Who are", or "What’s"
  // Fynbos is matched exactly, with an optional question mark at the end
  const pattern = /^(Who is|What is|Who are|What’s) Fynbos\??$/i;

  // Test the input string against the pattern
  if (pattern.test(term)) {
    // If the input matches, return "ladida"
    return `<p>Hi 👋 We're <strong>Fynbos</strong>, a global fintech company based in Cape Town solving complex problems with simple solutions that are a joy to use.</p>
    <p>To learn about what we’ve built, you can check out our <a href="/wallet">wallet</a> or <a href="/wealth">wealth</a> products.</p>`;
  }

  const pattern2 = /^(What have you built|What products do you have|What are your products)\??$/i;
  if (pattern2.test(term)) {
    return `<h2>Digital Wallet for the Interledger Foundation</h2>
    <p>We build a digital wallet for the Interledger Foundation that makes earning money online as easy as clicking a button. It’s got some great features like the ability to send money to people using their social handles as identifiers.</p>
    
    <h2>Upcoming Wealth Management Platform</h2>
    <p>We are busy building a new wealth management platform that we hope to release soon. It will make understanding saving and investment simple and accessible to anyone.</p>
    
    <p>You can test drive some of the early ideas and sign-up for the waitlist at <a href="https://wealth.fynbos.app">wealth.fynbos.app</a>.</p>`
  }

  const pattern3 = /^(How do I contact you|What’s your email address|Are you on (Twitter|Facebook|Instagram|Insta))\??$/i;
  if (pattern3.test(term)) {
    return `<p>You can fill out our contact form <a href="/contact">here</a> or send us an email at <a href="mailto:hello@fynbos.dev">hello@fynbos.dev</a>.</p>
    
    <p>We’re on social media at:</p>
    <ul>
        <li>YouTube: <a href="https://youtube.com/fynbosdev">youtube.com/fynbosdev</a></li>
        <li>Instagram: <a href="https://instagram.com/fynbosdev">instagram.com/fynbosdev</a></li>
        <li>Facebook: <a href="https://facebook.com/fynbosdev">facebook.com/fynbosdev</a></li>
        <li>TikTok: <a href="https://tiktok.com/@fynbosdev">@fynbosdev</a></li>
    </ul>`
  }

  // If the input doesn't match, return an empty string or any other indication of no match
  return "";
}

export async function autoNavigate(term :string) : Promise<string> {
  if (term.localeCompare("/about") == 0) {
    return "/about"
  }
  if (term.localeCompare("/wealth") == 0) {
    return "/wealth"
  }
  if (term.localeCompare("/wallet") == 0) {
    return "/wallet"
  }
  if (term.localeCompare("/legal") == 0) {
    return "/legal"
  }
  if (term.localeCompare("/contact") == 0) {
    return "/contact"
  }
  if (term.localeCompare("/blog") == 0) {
    return "/blog"
  }

  return ""
}
