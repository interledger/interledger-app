import { gql } from '@apollo/client'
import type { Query, QueryAllBlogPostsArgs } from '~/generated/dato-cms-graphql'
import { apolloClient } from '~/lib/apollo.server'
import type { QueryBlogPostArgs } from '~/generated/dato-cms-graphql'

// TODO possibly look into https://www.datocms.com/blog/how-to-generate-typescript-types-from-graphql#how-to-avoid-nullable-types-on-required-field-in-datocms

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

export const getAllBlogPosts = async (variables?: QueryAllBlogPostsArgs) => {
  return await apolloClient
    .query<{ allBlogPosts: Query['allBlogPosts'] }, QueryAllBlogPostsArgs>({
      query: gql`
        query GetAllBlogsQuery(
          $first: IntType
          $orderBy: [BlogPostModelOrderBy]
          $skip: IntType
        ) {
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
        }
      `,
      variables
    })
    .then((res) => {
      // console.log('DATA', res.data)
      return res.data.allBlogPosts
    })
    .catch((error) => {
      console.log(error)
    })
}

export const getCurrentBlogPost = async (variables: QueryBlogPostArgs) => {
  return await apolloClient
    .query<{ blogPost: Query['blogPost'] }, QueryBlogPostArgs>({
      query: gql`
        ${RESPONSIVE_IMAGE}
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
        }
      `,
      variables
    })
    .then((res) => {
      return res.data.blogPost
    })
    .catch((error) => {
      console.log(error)
    })
}
