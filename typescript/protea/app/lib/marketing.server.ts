import { gql } from '@apollo/client'
import type {
  Query,
  QueryAllBlogPostsArgs,
  QueryAllPeopleArgs,
  QueryBlogPostArgs,
  QueryDocArgs,
  QueryLegalPageArgs
} from '~/generated/dato-cms-graphql'
import { apolloClient } from '~/lib/apollo.server'

export const RESPONSIVE_IMAGE = gql`
  fragment ResponsiveImage on FileField {
    responsiveImage(imgixParams: { fit: max, auto: format }) {
      srcSet
      webpSrcSet
      sizes
      src
      width
      height
      aspectRatio
      alt
      title
      base64
    }
  }
`

export const getAboutRoute = async () => {
  return apolloClient
    .query<{
      aboutRoute: Query['aboutRoute']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetRouteContent {
          aboutRoute {
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
      return { aboutRoute: null, footer: null }
    })
}

export const getAllDocs = async () => {
  return apolloClient
    .query<{
      allDocs: Query['allDocs']
    }>({
      query: gql`
        query GetRouteContent {
          allDocs {
            id
            slug
            title
            sections {
              id
              title
              slug
            }
          }
        }
      `
    })
    .then((res) => {
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { allDocs: null }
    })
}

export const getCurrentDocPage = async (variables: QueryDocArgs) => {
  return apolloClient
    .query<{ doc: Query['doc']; footer: Query['footer'] }, QueryDocArgs>({
      query: gql`
        ${RESPONSIVE_IMAGE}
        ${FOOTER}
        query GetCurrentDocQuery($filter: DocModelFilter) {
          doc(filter: $filter) {
            id
            title
            slug
            _status
            sections {
              id
              title
              slug
              content {
                value
                blocks {
                  __typename
                  ... on InlineImageRecord {
                    id
                    altText
                    image {
                      ...ResponsiveImage
                    }
                    imageMobile {
                      ...ResponsiveImage
                    }
                    imageDark {
                      ...ResponsiveImage
                    }
                    imageDarkMobile {
                      ...ResponsiveImage
                    }
                  }
                  ... on InlineVideoRecord {
                    id
                    video {
                      provider
                      providerUid
                      thumbnailUrl
                      title
                      url
                    }
                  }
                }
              }
            }
            seoMeta: _seoMetaTags {
              tag
              attributes
              content
            }
            title
          }
          footer {
            ...Footer
          }
        }
      `,
      variables
    })
    .then((res) => {
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { doc: null, footer: null }
    })
}

export const getBlogRoute = async (variables?: QueryAllBlogPostsArgs) => {
  return apolloClient
    .query<
      {
        blogRoute: Query['blogRoute']
        allBlogPosts: Query['allBlogPosts']
        footer: Query['footer']
      },
      QueryAllBlogPostsArgs
    >({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetBlogRouteContent(
          $first: IntType
          $orderBy: [BlogPostModelOrderBy]
          $skip: IntType
        ) {
          blogRoute {
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
          allBlogPosts(first: $first, orderBy: $orderBy, skip: $skip) {
            id
            title
            slug
            description
            date
            _status
            shapes {
              url
            }
            shapesMobile {
              url
            }
            authors {
              name
            }
          }
          footer {
            ...Footer
          }
        }
      `,
      variables
    })
    .then((res) => {
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { blogRoute: null, allBlogPosts: null, footer: null }
    })
}

export const getCurrentBlogPost = async (variables: QueryBlogPostArgs) => {
  return apolloClient
    .query<
      { blogPost: Query['blogPost']; footer: Query['footer'] },
      QueryBlogPostArgs
    >({
      query: gql`
        ${RESPONSIVE_IMAGE}
        ${FOOTER}
        query GetCurrentBlogPostQuery($filter: BlogPostModelFilter) {
          blogPost(filter: $filter) {
            slug
            _status
            content {
              value
              blocks {
                __typename
                ... on InlineImageRecord {
                  id
                  altText
                  image {
                    ...ResponsiveImage
                  }
                  imageMobile {
                    ...ResponsiveImage
                  }
                  imageDark {
                    ...ResponsiveImage
                  }
                  imageDarkMobile {
                    ...ResponsiveImage
                  }
                }
                ... on InlineVideoRecord {
                  id
                  video {
                    provider
                    providerUid
                    thumbnailUrl
                    title
                    url
                  }
                }
                ... on InlinePersonRecord {
                  id
                  name
                  role
                  avatar {
                    responsiveImage(
                      imgixParams: { fit: max, w: 140, h: 140, auto: format }
                    ) {
                      srcSet
                      webpSrcSet
                      sizes
                      src
                      width
                      height
                      aspectRatio
                      alt
                      title
                      base64
                    }
                  }
                }
                ... on InlineTwitterEmbedRecord {
                  id
                  url
                  imageOfTweet {
                    ...ResponsiveImage
                  }
                }
              }
            }
            id
            authors {
              id
              name
              twitterUrl
              avatar {
                responsiveImage(
                  imgixParams: { fit: max, w: 80, h: 80, auto: format }
                ) {
                  srcSet
                  sizes
                  src
                  width
                  height
                  aspectRatio
                  alt
                  title
                  base64
                }
              }
            }
            shapes {
              url
            }
            shapesMobile {
              url
            }
            description
            date
            seoMeta: _seoMetaTags {
              tag
              attributes
              content
            }
            title
          }
          footer {
            ...Footer
          }
        }
      `,
      variables
    })
    .then((res) => {
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { blogPost: null, footer: null }
    })
}

export const getPerson = async (variables: QueryAllPeopleArgs) => {
  return apolloClient
    .query<{ person: Query['person'] }, QueryAllPeopleArgs>({
      query: gql`
        query GetPersonProfile($filter: PersonModelFilter) {
          person(filter: $filter) {
            id
            name
            avatar {
              responsiveImage(
                imgixParams: { fit: max, w: 120, h: 120, auto: format }
              ) {
                srcSet
                sizes
                src
                width
                height
                aspectRatio
                alt
                title
                base64
              }
            }
          }
        }
      `,
      variables
    })
    .then((res) => {
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { person: null }
    })
}

export const getContactRoute = async () => {
  return apolloClient
    .query<{
      contactRoute: Query['contactRoute']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetContactRouteContent {
          contactRoute {
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
      return { contactRoute: null, footer: null }
    })
}

export const getDiscordRoute = async () => {
  return apolloClient
    .query<{
      discordRoute: Query['discordRoute']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetDiscordRouteContent {
          discordRoute {
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
      return { discordRoute: null, footer: null }
    })
}
export const getSlackRoute = async () => {
  return apolloClient
    .query<{
      slackRoute: Query['slackRoute']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetSlackRouteContent {
          slackRoute {
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
      return { slackRoute: null, footer: null }
    })
}

export const getLegalRoute = async () => {
  return apolloClient
    .query<{
      legalRoute: Query['legalRoute']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetLegalRouteContent {
          legalRoute {
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
      return { legalRoute: null, footer: null }
    })
}

export const getCurrentLegalPage = async (variables: QueryLegalPageArgs) => {
  return apolloClient
    .query<
      { legalPage: Query['legalPage']; footer: Query['footer'] },
      QueryLegalPageArgs
    >({
      query: gql`
        ${FOOTER}
        query GetCurrentBlogPostQuery($filter: LegalPageModelFilter) {
          legalPage(filter: $filter) {
            slug
            external
            body {
              value
            }
            id
            title
            _publishedAt
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
      `,
      variables
    })
    .then((res) => {
      return res.data
    })
    .catch((error) => {
      console.log(error)
      return { legalPage: null, footer: null }
    })
}

export const getHomeRoute = async () => {
  return apolloClient
    .query<{
      homeRoute: Query['homeRoute']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetHomeRouteContent {
          homeRoute {
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
      return { homeRoute: null, footer: null }
    })
}

export const getWalletRoute = async () => {
  return apolloClient
    .query<{
      walletRoute: Query['walletRoute']
      footer: Query['footer']
    }>({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetWalletRouteContent {
          walletRoute {
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
      return { walletRoute: null, footer: null }
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
        textStandard
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
