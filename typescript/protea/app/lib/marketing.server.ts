import { gql } from '@apollo/client'
import type { Query } from '~/generated/dato-cms-graphql'
import { apolloClient } from '~/lib/apollo.server'
import { RESPONSIVE_IMAGE } from '~/lib/blog.server'

export const getAboutPage = async () => {
  return apolloClient
    .query<{
      aboutpage: Query['aboutpage']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetTestPageContent {
          aboutpage {
            id
            body {
              ...Section
            }
            _status
            seoMeta: _seoMetaTags {
              tag
              attributes
              content
            }
          }
          footer {
            ...Footer
          }
        }
      `
    })
    .then((res) => {
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { aboutpage: null, footer: null }
    })
}

export const getBlogPage = async () => {
  return apolloClient
    .query<{
      blogpage: Query['blogpage']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetTestPageContent {
          blogpage {
            id
            body {
              ...Section
            }
            _status
            seoMeta: _seoMetaTags {
              tag
              attributes
              content
            }
          }
          footer {
            ...Footer
          }
        }
      `
    })
    .then((res) => {
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { blogpage: null, footer: null }
    })
}

export const getHomePage = async () => {
  return apolloClient
    .query<{
      homepage: Query['homepage']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetTestPageContent {
          homepage {
            id
            body {
              ...Section
            }
            _status
            seoMeta: _seoMetaTags {
              tag
              attributes
              content
            }
          }
          footer {
            ...Footer
          }
        }
      `
    })
    .then((res) => {
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { homepage: null, footer: null }
    })
}

export const getWalletPage = async () => {
  return apolloClient
    .query<{
      walletpage: Query['walletpage']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetPageContent {
          walletpage {
            id
            body {
              ...Section
            }
            _status
            seoMeta: _seoMetaTags {
              tag
              attributes
              content
            }
          }
          footer {
            ...Footer
          }
        }
      `
    })
    .then((res) => {
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { walletpage: null, footer: null }
    })
}

export const FOOTER = gql`
  fragment Footer on FooterRecord {
    id
    logo {
      id
      url
    }
    column1Title
    column1 {
      id
      displayText
      url
    }
    column2Title
    column2 {
      id
      displayText
      url
    }
    column3Title
    column3 {
      id
      displayText
      url
    }
    legalText {
      value
    }
    socialIcons {
      id
      url
      icon {
        id
        url
      }
    }
  }
`

export const SECTION = gql`
  ${RESPONSIVE_IMAGE}
  fragment Section on SectionRecord {
    id
    content {
      ... on CtaContentRecord {
        title
        body
        button {
          id
          displayText
          url
          button
        }
        image {
          id
          url
        }
        imageDark {
          id
          url
        }
      }
      ... on FeatureBlocksContentRecord {
        blocks {
          id
          image {
            id
            url
          }
          title
          direction
          backgroundColour {
            hex
          }
        }
      }
      ... on FeatureContentRecord {
        title
        body
        image {
          id
          url
        }
        imageMobile {
          id
          url
        }
        rowReverse
      }
      ... on HeaderContentRecord {
        id
        title
        shapes {
          id
          url
        }
      }
      ... on HeroContentRecord {
        id
        title
        body
        button {
          id
          displayText
          url
          button
        }
        image {
          id
          url
        }
        imageDark {
          id
          url
        }
        imageMobile {
          id
          url
        }
        imageDarkMobile {
          id
          url
        }
      }
      ... on HomeHeroContentRecord {
        id
        title
        body
        button {
          id
          displayText
          url
          button
        }
        iterations {
          id
          title
          body
          image {
            id
            url
          }
          imageDark {
            id
            url
          }
          imageMobile {
            id
            url
          }
          imageDarkMobile {
            id
            url
          }
          mobileShape {
            id
            url
          }
        }
      }
      ... on ShowcaseContentRecord {
        id
        rowReverse
        cases {
          id
          title
          body
          image {
            id
            url
          }
          imageDark {
            id
            url
          }
        }
      }
      ... on StoryContentRecord {
        id
        title
        blurb {
          value
        }
        bodyText {
          value
        }
        image {
          id
          url
        }
        imageDark {
          id
          url
        }
      }
      ... on TeamContentRecord {
        id
        title
        image {
          id
          url
        }
        imageDark {
          id
          url
        }
        people {
          id
          person {
            name
            role
            twitterUrl
            linkedinUrl
            fynbosUrl
            avatar {
              ...ResponsiveImage
            }
          }
        }
      }
      ... on TextContentRecord {
        id
        title
        bodyText {
          value
        }
        textCentered
        button {
          id
          displayText
          url
          button
        }
      }
    }
  }
`
