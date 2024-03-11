import { gql } from '@apollo/client'
import type {
  Query,
  QueryAllBlogPostsArgs,
  QueryBlogPostArgs,
  QueryDocArgs,
  QueryLegalPageArgs,
  QueryMarketingPageArgs
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
            _seoMetaTags {
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
            _seoMetaTags {
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
              height
              width
            }
            shapesMobile {
              url
              height
              width
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
              height
              width
            }
            shapesMobile {
              url
              height
              width
            }
            description
            date
            _seoMetaTags {
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
            _seoMetaTags {
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

export const getCurrentMarketingPage = async (
  variables: QueryMarketingPageArgs
) => {
  return apolloClient
    .query<
      { marketingPage: Query['marketingPage']; footer: Query['footer'] },
      QueryMarketingPageArgs
    >({
      query: gql`
        ${FOOTER}
        ${SECTION}
        query GetCurrentMarketingPageQuery($filter: MarketingPageModelFilter) {
          marketingPage(filter: $filter) {
            slug
            body {
              ...Section
            }
            id
            title
            _publishedAt
            _seoMetaTags {
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
      return { marketingPage: null, footer: null }
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
            _seoMetaTags {
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
            _seoMetaTags {
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

export const FOOTER = gql`
  fragment Footer on FooterRecord {
    id
    logo {
      id
      url
      height
      width
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
        height
        width
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
          height
          width
        }
        imageDark {
          id
          url
          height
          width
        }
      }
      ... on FeatureBlocksContentRecord {
        blocks {
          id
          image {
            id
            url
            height
            width
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
          height
          width
        }
        imageMobile {
          id
          url
          height
          width
        }
        rowReverse
      }
      ... on HeaderContentRecord {
        id
        title
        shapes {
          id
          url
          height
          width
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
          height
          width
        }
        imageDark {
          id
          url
          height
          width
        }
        imageMobile {
          id
          url
          height
          width
        }
        imageDarkMobile {
          id
          url
          height
          width
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
            height
            width
          }
          imageDark {
            id
            url
            height
            width
          }
          imageMobile {
            id
            url
            height
            width
          }
          imageDarkMobile {
            id
            url
            height
            width
          }
          mobileShape {
            id
            url
            height
            width
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
            height
            width
          }
          imageDark {
            id
            url
            height
            width
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
          height
          width
        }
        imageDark {
          id
          url
          height
          width
        }
      }
      ... on TeamContentRecord {
        id
        title
        image {
          id
          url
          height
          width
        }
        imageDark {
          id
          url
          height
          width
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
        image {
          height
          width
          url
        }
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
