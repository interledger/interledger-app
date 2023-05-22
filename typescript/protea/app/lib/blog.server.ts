import { gql } from '@apollo/client'
import type { Query, QueryAllBlogPostsArgs } from '~/generated/dato-cms-graphql'
import { apolloClient } from '~/lib/apollo.server'
import type { QueryBlogPostArgs } from '~/generated/dato-cms-graphql'

// TODO possibly look into https://www.datocms.com/blog/how-to-generate-typescript-types-from-graphql#how-to-avoid-nullable-types-on-required-field-in-datocms

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
              responsiveImage(
                imgixParams: { fit: max, w: 120, h: 120, auto: format }
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
                  image {
                    url
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
                ... on InlineTwitterEmbedRecord {
                  id
                  url
                  imageOfTweet {
                    url
                    alt
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
                url
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
            seoMeta {
              twitterCard
              title
              description
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
