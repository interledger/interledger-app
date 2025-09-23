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
  return {
    contactRoute: {
      id: '125075088',
      body: [
        {
          __typename: 'SectionRecord',
          id: '125075082',
          content: [
            {
              id: '125075081',
              title: 'Contact',
              body: 'Get in touch, we’re here to help.',
              button: [],
              image: {
                id: 'eQBZnAgpS5m8UyieKh_7dQ',
                url: 'https://www.datocms-assets.com/160242/1710165703-contact-hero-light.svg',
                height: 680,
                width: 680,
                __typename: 'FileField'
              },
              imageDark: {
                id: 'eQBZnAgpS5m8UyieKh_7dQ',
                url: 'https://www.datocms-assets.com/160242/1710165703-contact-hero-light.svg',
                height: 680,
                width: 680,
                __typename: 'FileField'
              },
              imageMobile: {
                id: '51207375',
                url: 'https://www.datocms-assets.com/160242/1685717318-hero-5-contact-light-mobile.svg',
                height: 187,
                width: 375,
                __typename: 'FileField'
              },
              imageDarkMobile: {
                id: '51207373',
                url: 'https://www.datocms-assets.com/160242/1685717314-hero-5-contact-dark-mobile.svg',
                height: 187,
                width: 375,
                __typename: 'FileField'
              },
              __typename: 'HeroContentRecord'
            }
          ]
        },
        {
          __typename: 'SectionRecord',
          id: '125075084',
          content: [{ __typename: 'ManualContentRecord' }]
        }
      ],
      _status: 'published',
      _seoMetaTags: [
        {
          tag: 'title',
          attributes: null,
          content: 'Interledger wallet',
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: { property: 'og:title', content: 'Interledger wallet' },
          content: null,
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: { name: 'twitter:title', content: 'Interledger wallet' },
          content: null,
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: { name: 'description', content: '.' },
          content: null,
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: { property: 'og:description', content: '.' },
          content: null,
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: { name: 'twitter:description', content: '.' },
          content: null,
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: { property: 'og:locale', content: 'en' },
          content: null,
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: { property: 'og:type', content: 'website' },
          content: null,
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: { property: 'og:site_name', content: 'Interledger' },
          content: null,
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: {
            property: 'article:modified_time',
            content: '2024-03-11T14:02:15Z'
          },
          content: null,
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: { property: 'article:publisher', content: '' },
          content: null,
          __typename: 'Tag'
        },
        {
          tag: 'meta',
          attributes: { name: 'twitter:card', content: 'summary_large_image' },
          content: null,
          __typename: 'Tag'
        }
      ],
      __typename: 'ContactRouteRecord'
    } as Query['contactRoute'],
    footer: {
      __typename: 'FooterRecord',
      id: '121270040',
      logo: {
        id: 'dffIomSsTfCkd-b3vjChtA',
        url: '/interledger-logo.svg',
        height: 150,
        width: 150,
        __typename: 'FileField'
      },
      column1Title: 'Menu',
      column1: [
        {
          id: '125075096',
          displayText: 'Contact',
          url: 'https://interledger.app/contact',
          __typename: 'LinkRecord'
        }
      ],
      column2Title: 'Ecosystem',
      column2: [
        {
          id: '121270066',
          displayText: 'Interledger Foundation',
          url: 'https://interledger.org/',
          __typename: 'LinkRecord'
        },
        {
          id: '125075093',
          displayText: 'Web monetization',
          url: 'https://webmonetization.org/',
          __typename: 'LinkRecord'
        },
        {
          id: '125075094',
          displayText: 'Open Payments',
          url: 'https://openpayments.dev/',
          __typename: 'LinkRecord'
        }
      ],
      column3Title: 'Legal',
      column3: [
        {
          id: '121270067',
          displayText: 'Legal Agreements',
          url: 'https://interledger.app/legal',
          __typename: 'LinkRecord'
        }
      ],
      legalText: {
        value: {
          schema: 'dast',
          document: {
            type: 'root',
            children: [
              {
                type: 'paragraph',
                children: [
                  {
                    type: 'span',
                    value:
                      '© 2024 Interledger Inc and the Interledger Foundation.'
                  }
                ]
              },
              {
                type: 'paragraph',
                children: [
                  {
                    type: 'span',
                    value:
                      'The Interledger name and logo are the property of the'
                  },
                  {
                    url: 'https://interledger.org',
                    type: 'link',
                    children: [
                      { type: 'span', value: ' Interledger Foundation' }
                    ]
                  },
                  {
                    type: 'span',
                    value:
                      '. The Interledger app is powered by Interledger on behalf of the Interledger Foundation as a service to the Interledger community. Interledger Inc is not a bank. Interledger provides a technology platform and all payments and banking services are provided by our partners who are appropriately licensed.'
                  }
                ]
              }
            ]
          }
        },
        __typename: 'FooterModelLegalTextField'
      },
      socialIcons: [
        {
          id: '124003009',
          url: 'https://x.com/Interledger',
          icon: {
            id: 'If7vlze6QAeqjixiyS922g',
            url: 'https://www.datocms-assets.com/160242/1710406464-x-white-icon.svg',
            height: 24,
            width: 24,
            __typename: 'FileField'
          },
          __typename: 'SocialIconRecord'
        },
        {
          id: '126089438',
          url: 'https://www.linkedin.com/company/interledger-foundation/',
          icon: {
            id: '52090585',
            url: 'https://www.datocms-assets.com/160242/1685976901-icon-linkedin.svg',
            height: 16,
            width: 16,
            __typename: 'FileField'
          },
          __typename: 'SocialIconRecord'
        },
        {
          id: '126089519',
          url: 'https://www.youtube.com/@InterledgerFoundation',
          icon: {
            id: '52090589',
            url: 'https://www.datocms-assets.com/160242/1685977020-icon-youtube.svg',
            height: 13,
            width: 19,
            __typename: 'FileField'
          },
          __typename: 'SocialIconRecord'
        },
        {
          id: 'PCadAp0bSRW4_0v6BTeiKg',
          url: 'https://www.instagram.com/interledgerfoundation/',
          icon: {
            id: 'dfWXhnzmQT68juUEGsvCkw',
            url: 'https://www.datocms-assets.com/160242/1710227274-instagram-white-icon.svg',
            height: 24,
            width: 24,
            __typename: 'FileField'
          },
          __typename: 'SocialIconRecord'
        },
        {
          id: 'JFxCqHyWQCyJbpfrhMbeJQ',
          url: 'https://www.facebook.com/interledger',
          icon: {
            id: 'YhE1LNyHRSWKFCxYuStcYg',
            url: 'https://www.datocms-assets.com/160242/1710227315-facebook-white-icon.svg',
            height: 24,
            width: 24,
            __typename: 'FileField'
          },
          __typename: 'SocialIconRecord'
        }
      ]
    } as Query['footer']
  }
}

export const getCurrentMarketingPage = async (
  variables: QueryMarketingPageArgs
) => {
  switch (variables?.filter?.slug?.eq) {
    case 'legal': {
      return {
        marketingPage: {
          slug: 'legal',
          body: [
            {
              __typename: 'SectionRecord',
              id: 'CXMaL5P2QbeSvuoZR7vmdw',
              content: [
                {
                  id: 'Y40mIGslRhK3YAxqr7geRg',
                  title: 'Legal agreements',
                  body: 'The Interledger app is available to users around the world subject to different regulations and legal requirements. Here is a complete list of policies, agreements and disclosures.',
                  button: [],
                  image: {
                    id: 'RTDot43vSvuc_m_O8Mhw_Q',
                    url: 'https://www.datocms-assets.com/160242/1710166277-legal-hero-light.svg',
                    height: 680,
                    width: 680,
                    __typename: 'FileField'
                  },
                  imageDark: {
                    id: 'AMhOYFBjTvCAHD2MYekP8w',
                    url: 'https://www.datocms-assets.com/160242/1710166264-legal-hero-dark.svg',
                    height: 680,
                    width: 680,
                    __typename: 'FileField'
                  },
                  imageMobile: {
                    id: 'HmPLp8bASl2u4gHLO7Xonw',
                    url: 'https://www.datocms-assets.com/160242/1710157616-legal-mobile-hero-light.svg',
                    height: 187,
                    width: 375,
                    __typename: 'FileField'
                  },
                  imageDarkMobile: {
                    id: 'Rs1LF2lTQTiUGGV0JVxkkA',
                    url: 'https://www.datocms-assets.com/160242/1710157628-legal-mobile-hero-dark.svg',
                    height: 187,
                    width: 375,
                    __typename: 'FileField'
                  },
                  __typename: 'HeroContentRecord'
                }
              ]
            },
            {
              __typename: 'SectionRecord',
              id: 'Thgapr7PTjKxZq1O7bdC5Q',
              content: [
                {
                  id: 'QxTc5bpxT2G8eFjWCqFXMQ',
                  title: '',
                  image: null,
                  bodyText: {
                    value: {
                      schema: 'dast',
                      document: {
                        type: 'root',
                        children: [
                          {
                            type: 'heading',
                            level: 2,
                            children: [{ type: 'span', value: 'For all users' }]
                          },
                          {
                            type: 'list',
                            style: 'bulleted',
                            children: [
                              {
                                type: 'listItem',
                                children: [
                                  {
                                    type: 'paragraph',
                                    children: [
                                      {
                                        url: 'https://interledger.app/legal/terms-of-service',
                                        type: 'link',
                                        children: [
                                          {
                                            type: 'span',
                                            value: 'Terms of Service'
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              },
                              {
                                type: 'listItem',
                                children: [
                                  {
                                    type: 'paragraph',
                                    children: [
                                      {
                                        url: 'https://interledger.app/legal/privacy-policy',
                                        type: 'link',
                                        children: [
                                          {
                                            type: 'span',
                                            value: 'Privacy Policy'
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              },
                              {
                                type: 'listItem',
                                children: [
                                  {
                                    type: 'paragraph',
                                    children: [
                                      {
                                        url: 'https://interledger.app/legal/wallet-license',
                                        type: 'link',
                                        children: [
                                          {
                                            type: 'span',
                                            value: 'Wallet Address License'
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              },
                              {
                                type: 'listItem',
                                children: [
                                  {
                                    type: 'paragraph',
                                    children: [
                                      {
                                        url: 'https://interledger.app/legal/accessibility-statement',
                                        type: 'link',
                                        children: [
                                          {
                                            type: 'span',
                                            value: 'Accessibility Statement'
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              }
                            ]
                          },
                          {
                            type: 'paragraph',
                            children: [{ type: 'span', value: '' }]
                          },
                          {
                            type: 'heading',
                            level: 3,
                            children: [{ type: 'span', value: 'For US users' }]
                          },
                          {
                            type: 'list',
                            style: 'bulleted',
                            children: [
                              {
                                type: 'listItem',
                                children: [
                                  {
                                    type: 'paragraph',
                                    children: [
                                      {
                                        url: 'https://interledger.app/legal/us/e-sign-agreement',
                                        type: 'link',
                                        children: [
                                          {
                                            type: 'span',
                                            value: 'eSign Agreement'
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              },
                              {
                                type: 'listItem',
                                children: [
                                  {
                                    type: 'paragraph',
                                    children: [
                                      {
                                        url: 'https://astrafi.com/terms/',
                                        type: 'link',
                                        children: [
                                          {
                                            type: 'span',
                                            value: 'Astra Terms of Use'
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              },
                              {
                                type: 'listItem',
                                children: [
                                  {
                                    type: 'paragraph',
                                    children: [
                                      {
                                        url: 'https://astrafi.com/privacy/',
                                        type: 'link',
                                        children: [
                                          {
                                            type: 'span',
                                            value: 'Astra Privacy Policy'
                                          }
                                        ]
                                      }
                                    ]
                                  }
                                ]
                              }
                            ]
                          }
                        ]
                      }
                    },
                    __typename: 'TextContentModelBodyTextField'
                  },
                  textCentered: false,
                  textStandard: true,
                  button: [],
                  __typename: 'TextContentRecord'
                }
              ]
            }
          ],
          id: 'K1OqA06kQ-2XP266ixIF7w',
          title: 'Legal agreements',
          _publishedAt: '2024-11-20T14:17:13+02:00',
          _seoMetaTags: [
            {
              tag: 'title',
              attributes: null,
              content: 'Legal agreements',
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:title', content: 'Legal agreements' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                name: 'twitter:title',
                content: 'Legal agreements'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'twitter:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:locale', content: 'en' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:type', content: 'article' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:site_name', content: 'Interledger' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                property: 'article:modified_time',
                content: '2024-11-20T12:17:11Z'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'article:publisher', content: '' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                name: 'twitter:card',
                content: 'summary_large_image'
              },
              content: null,
              __typename: 'Tag'
            }
          ],
          __typename: 'MarketingPageRecord'
        } as unknown as Query['marketingPage'],
        footer: {
          __typename: 'FooterRecord',
          id: '121270040',
          logo: {
            id: 'dffIomSsTfCkd-b3vjChtA',
            url: 'https://www.datocms-assets.com/160242/1721916494-interledger_icon.svg',
            height: 101,
            width: 101,
            __typename: 'FileField'
          },
          column1Title: 'Menu',
          column1: [
            {
              id: '125075096',
              displayText: 'Contact',
              url: 'https://interledger.app/contact',
              __typename: 'LinkRecord'
            }
          ],
          column2Title: 'Ecosystem',
          column2: [
            {
              id: '121270066',
              displayText: 'Interledger Foundation',
              url: 'https://interledger.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075093',
              displayText: 'Web monetization',
              url: 'https://webmonetization.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075094',
              displayText: 'Open Payments',
              url: 'https://openpayments.dev/',
              __typename: 'LinkRecord'
            }
          ],
          column3Title: 'Legal',
          column3: [
            {
              id: '121270067',
              displayText: 'Legal Agreements',
              url: 'https://interledger.app/legal',
              __typename: 'LinkRecord'
            }
          ],
          legalText: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          '© 2024 Interledger Inc and the Interledger Foundation.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'The Interledger name and logo are the property of the'
                      },
                      {
                        url: 'https://interledger.org',
                        type: 'link',
                        children: [
                          { type: 'span', value: ' Interledger Foundation' }
                        ]
                      },
                      {
                        type: 'span',
                        value:
                          '. The Interledger app is powered by Interledger on behalf of the Interledger Foundation as a service to the Interledger community. Interledger Inc is not a bank. Interledger provides a technology platform and all payments and banking services are provided by our partners who are appropriately licensed.'
                      }
                    ]
                  }
                ]
              }
            },
            __typename: 'FooterModelLegalTextField'
          },
          socialIcons: [
            {
              id: '124003009',
              url: 'https://x.com/Interledger',
              icon: {
                id: 'If7vlze6QAeqjixiyS922g',
                url: 'https://www.datocms-assets.com/160242/1710406464-x-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089438',
              url: 'https://www.linkedin.com/company/interledger-foundation/',
              icon: {
                id: '52090585',
                url: 'https://www.datocms-assets.com/160242/1685976901-icon-linkedin.svg',
                height: 16,
                width: 16,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089519',
              url: 'https://www.youtube.com/@InterledgerFoundation',
              icon: {
                id: '52090589',
                url: 'https://www.datocms-assets.com/160242/1685977020-icon-youtube.svg',
                height: 13,
                width: 19,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'PCadAp0bSRW4_0v6BTeiKg',
              url: 'https://www.instagram.com/interledgerfoundation/',
              icon: {
                id: 'dfWXhnzmQT68juUEGsvCkw',
                url: 'https://www.datocms-assets.com/160242/1710227274-instagram-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'JFxCqHyWQCyJbpfrhMbeJQ',
              url: 'https://www.facebook.com/interledger',
              icon: {
                id: 'YhE1LNyHRSWKFCxYuStcYg',
                url: 'https://www.datocms-assets.com/160242/1710227315-facebook-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            }
          ]
        } as Query['footer']
      }
    }
    default: {
      return {
        marketingPage: null,
        footer: null
      }
    }
  }
}

export const getCurrentLegalPage = async (variables: QueryLegalPageArgs) => {
  switch (variables?.filter?.slug?.eq) {
    case 'terms-of-service':
      return {
        legalPage: {
          slug: 'terms-of-service',
          external: '',
          body: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      { type: 'span', value: 'These Terms of Service (“' },
                      { type: 'span', marks: ['emphasis'], value: 'Terms' },
                      {
                        type: 'span',
                        value:
                          '") apply to your access to and use of the Services (as defined in Section 1 below) provided by of the Interledger Foundation, a California nonprofit public benefit corporation (the "Corporation"). '
                      },
                      {
                        type: 'span',
                        marks: ['strong'],
                        value:
                          'By clicking “I Accept” or by accessing or using our Services, you agree to these Terms, including the mandatory arbitration provision and class action waiver in Section 17 and the provisions relating to future modification, termination and migration of our Services in Section 19. If you do not agree to these Terms, do not use our Services.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'If you have any questions about these Terms or our Services, please contact us at '
                      },
                      {
                        url: 'mailto:support@interledger.app',
                        type: 'link',
                        children: [
                          { type: 'span', value: 'support@interledger.app' }
                        ]
                      },
                      {
                        type: 'span',
                        value:
                          '. For information about how we collect, use, share and otherwise process information about you, please see our Privacy Policy.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value: 'You and Corporation agree as follows:‍'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [{ type: 'span', value: '1. Overview and Scope' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Corporation has developed a digital wallet that is accessible through the Corporation\'s website (the "'
                      },
                      { type: 'span', marks: ['emphasis'], value: 'Platform' },
                      { type: 'span', value: '”) at https://interledger.app. ' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'People who use the Services may license a unique wallet address, a URL, from the Corporation and associate it on the Platform with a third-party website URL, social media handle, payment account, email address, location data, text data, or other online data selected by the user (the "'
                      },
                      {
                        type: 'span',
                        marks: ['emphasis'],
                        value: 'Associated Data'
                      },
                      {
                        type: 'span',
                        value: '”) in accordance with these Terms.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value: 'These Terms govern the services (the “'
                      },
                      { type: 'span', marks: ['emphasis'], value: 'Services' },
                      {
                        type: 'span',
                        value:
                          '") that Corporation makes available to Platform users, including use or access to the Platform and other services provided by Corporation, the ability to license and associate a wallet address with Associated Data, resolve queries of a wallet address to the Associated Data, and obtain related services through the Platform. All references to Services in these Terms include wallet addresses unless otherwise specified.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [{ type: 'span', value: '2. Eligibility' }]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You must be at least 18 years of age to use our Services. If you are under 18 years of age (or the age of legal majority where you live), you may use our Services only under the supervision of a parent or legal guardian who agrees to be bound by these Terms. If you are a parent or legal guardian of a user under the age of 18 (or the age of legal majority), you agree to be fully responsible for the acts or omissions of such user in relation to our Services. If you use our Services on behalf of another person or entity, (i) all references to “you” throughout these Terms will include that person or entity; (ii) you represent that you are authorised to accept these Terms on that person’s or entity’s behalf; and (iii) in the event you or the person or entity violates these Terms, the person or entity agrees to be responsible to us.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may not access or use our Services if:'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'bulleted',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'you have been suspended from using our Services;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'under the applicable law of the jurisdiction(s) in which you reside or conduct business, you are prohibited from using the Services or do not have the requisite licenses or other governmental authorizations to use the Services;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'you are located in a country that is subject to a U.S. government embargo or that has been designated by the U.S. government as a “terrorist supporting” country;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'you are listed on any U.S. government list of prohibited or restricted parties; or'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'your use of the Services breaches any other agreement to which you are a party.‍'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      {
                        type: 'span',
                        value: '3. User Accounts and Account Security'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You need to register for an account to access some or all of our Services, including to register a wallet address and associate your wallet with Associated Data. If you register for an account, you must provide accurate account information and promptly update this information if it changes. You are responsible for the activities that occur in connection with your account and must maintain the security of your account. You are prohibited from sharing your password or other log-in credentials with any other person. Promptly notify us if you discover or suspect that someone has accessed your account without your permission.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'We can terminate or suspend your account at any time at our discretion. We are not responsible for any loss or harm related to your inability to access or use the Services. You may not bring a claim against us for suspending or terminating another person’s account, and you agree you will not bring such a claim. If you try to bring such a claim, you are responsible for the damages caused, including attorneys’ fees and costs.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You agree that your account is not transferable and that in the event of your death, incapacity or unavailability, we may terminate any rights to your account and wallet addresses.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      { type: 'span', value: '4. Registering a wallet address' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'This Section 4 applies if you activate a wallet address as part of our Services.'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'To register and activate a wallet address, you need to successfully complete a Know Your Customer (KYC) process through your account. You agree to pay all fees due for a wallet address at the time you activate the address, according to Corporation’s current fee schedule.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Corporation may modify its fee schedule at any time, and modifications will be posted on the Platform and effective immediately with respect to future registrations without further notice. All registrations will be subject to these Terms, including the limited license set forth in Section 7, and will be non-refundable unless otherwise agreed.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Corporation may collect a fee and/or commission (each, a “Transaction Fee”) on the total value of any transaction you undertake with your wallet. The applicable Transaction Fee will be determined by Corporation from time to time in its sole discretion and, if applicable, will be communicated to wallet owners at least 30 days prior to going into effect.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Except as required by law, all purchases are final and non-refundable. No refunds, returns or exchanges will be permitted for any reason. ALL SALES ARE FINAL.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [{ type: 'span', value: '5. Associated Data' }]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Our Services allow you and other users to associate a wallet with Associated Data through your account. Except for the license you grant below, you retain all rights in and to the Associated Data that you associate with your wallet, as between you and Corporation.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You grant Corporation and its subsidiaries and affiliates a nonexclusive, royalty-free, worldwide, fully paid, and sublicensable license to use, reproduce, modify, adapt, publish, translate, create derivative works from, distribute, publicly perform and display your Associated Data and any name, username or likeness provided in connection with your Associated Data in all media formats and channels now known or later developed without compensation to you. When you associate or otherwise share your Associated Data on or through our Services, you understand that your Associated Data may be visible to others.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may not associate or otherwise share any Associated Data that violates these Terms, that is confidential, or for which you do not have all the rights necessary to disclose and to grant us the license described above. In addition to the requirements in Section 6(b) below, you represent and warrant that your Associated Data, and our use of such Associated Data as permitted by these Terms, will not violate any rights of or cause injury to any person or entity. Although we have no obligation to screen, edit or monitor Associated Data, we may delete or remove your Associated Data at any time and for any reason with or without notice.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      {
                        type: 'span',
                        value: '6. Prohibited Conduct and Associated Data'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You are solely responsible for your conduct while using or accessing our Services. While using or accessing our Services, you will not:'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'bulleted',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Violate any applicable law, contract, intellectual property right or other third-party right or commit a tort;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Use our Services or for any illegal or unauthorized purpose, or engage in, encourage or promote any activity that violates these Terms;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Attempt to circumvent any content-limiting techniques we employ;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Use or attempt to use another user’s account unless authorized to do so by that user and Corporation;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Impersonate or post Associated Data on behalf of any person or entity or otherwise misrepresent your affiliation with a person or entity;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Use our Services other than for their intended purpose and in any manner that could interfere with, disrupt, negatively affect or inhibit other users from fully enjoying our Services or that could damage, disable, overburden or impair the functioning of our Services in any manner;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Reverse engineer any aspect of our Services or do anything that might discover source code or bypass or circumvent measures employed to prevent or limit access to any part of our Services;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Use any data mining, robots or similar data gathering or extraction methods designed to scrape or extract data from our Services;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Develop or use any applications that interact with our Services without our prior written consent;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Use our Services for benchmarking purposes or for the purpose of developing a competitive product;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Manipulate, or attempt to manipulate, our Services in any way;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Send, distribute or post spam, unsolicited or bulk commercial electronic communications, chain letters, or pyramid schemes;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Bypass or ignore instructions contained in our robots.txt file;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Engage in any harassing, threatening, intimidating, predatory or stalking conduct; or'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Engage in any conduct that, in our sole judgment, is objectionable, restricts or inhibits any other person from using or enjoying our Services, or may expose Corporation or others to any harm or liability of any type.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may not associate any Associated Data with your wallet that:'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'bulleted',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Is unlawfully provided, however, we neither review nor evaluate the content hosted by the third party websites (“'
                              },
                              {
                                type: 'span',
                                marks: ['emphasis'],
                                value: 'Third Party Sites'
                              },
                              {
                                type: 'span',
                                value:
                                  '”) whose URLs you may associate with your wallet and assume no liability or responsibility for the content of such Third Party Sites;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Would constitute, encourage or provide instructions for a criminal offense, violate the rights of any party or otherwise create liability or violate any law;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'May infringe any patent, trademark, trade secret, copyright or other intellectual or proprietary right of any party;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Contains or depicts any statements, remarks or claims that do not reflect your honest views and experiences;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Impersonates, or misrepresents your affiliation with, any person or entity;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Contains any unsolicited promotions, political campaigning, advertising or solicitations;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Contains any private or personal information of a third party without such third party’s consent;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Contains any viruses, corrupted data or other harmful, disruptive or destructive files or content; or'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'In our sole judgment, is objectionable, restricts or inhibits any other person from using or enjoying our Services, or may expose Corporation or others to any harm or liability of any type.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  "Enforcement of this Section 6 is solely at Corporation's discretion, and failure to enforce this section in some instances does not constitute a waiver of our right to enforce it in other instances. In addition, this Section 6 does not create any private right of action on the part of any third party or any reasonable expectation that the Services will not contain any content that is prohibited by such rules."
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      { type: 'span', value: '7. Ownership; Limited Licenses' }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The Services (which includes and any constituent part thereof including any wallet address, text, graphics, images, photographs, videos, illustrations and other content contained therein), are owned by Corporation or our licensors and are protected under both United States and foreign laws. Except as explicitly stated in these Terms, all rights in and to the Services are reserved by us or our licensors.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Subject to your compliance with these Terms, you are hereby granted a limited, nonexclusive, non-sublicensable, non-transferable, revocable license to access and use our Services for your own personal, noncommercial use.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'If you register a wallet address in accordance with Section 4, and subject to your compliance with these Terms, you are hereby granted a limited, exclusive, non-sublicensable, non-transferable, revocable license to access, copy, display, distribute and use the Services solely for the purpose of allowing third-parties to perform lookups of Associated Data associated to such wallet address via the Corporation’s application programming interface.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Without limiting the foregoing provisions in this Section 7, you will not, directly or indirectly: (i) modify or create derivative works of the Services in whole or in part; (ii) rent, lease, lend, sell, advertise, assign, encumber, or otherwise commercially use the Services; (iii) remove any proprietary notices from the Services; or (iv) use the Services in any manner or for any purpose that infringes, misappropriates, or otherwise violates any intellectual property right or other right of the Corporation or any other person, or that violates any applicable law.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Any use of the Services other than as specifically authorized herein, without our prior written permission, is (i) strictly prohibited; (ii) will immediately terminate the license for Services in Section 7.2 or the license for a wallet address in Section 7.3, as applicable; and (iii) violate our intellectual property rights. If your wallet address license is terminated, you will immediately lose access to your wallet.‍'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [{ type: 'span', value: '8. Trademarks' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Interledger Inc, the Interledger mark and our logos, our product or service names, our slogans and the look and feel of the Services are trademarks of Corporation and may not be copied, imitated or used, in whole or in part, without our prior written permission. All other trademarks, registered trademarks, product names and company names or logos mentioned on the Services are the property of their respective owners. Reference to any products, services, processes or other information by trade name, trademark, manufacturer, supplier or otherwise does not constitute or imply endorsement, sponsorship or recommendation by us.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [{ type: 'span', value: '9. Feedback' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'You may voluntarily post, submit or otherwise communicate to us any questions, comments, suggestions, ideas, original or creative materials or other information about Corporation or our Services (collectively, “'
                      },
                      { type: 'span', marks: ['emphasis'], value: 'Feedback' },
                      {
                        type: 'span',
                        value:
                          '”). You understand that we may use such Feedback for any purpose, commercial or otherwise, without acknowledgment or compensation to you, including to develop, copy, publish, or improve the Feedback in Corporation’s sole discretion. You understand that Corporation may treat Feedback as nonconfidential.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      {
                        type: 'span',
                        value:
                          '10. Repeat Infringer Policy; Copyright Complaints'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'In accordance with the Digital Millennium Copyright Act and other applicable law, we have adopted a policy of terminating, in appropriate circumstances, the accounts of users who repeatedly infringe the intellectual property rights of others. If you believe that anything on our Services infringes any copyright that you own or control, you may notify Corporation’s designated agent as follows by sending an email to '
                      },
                      {
                        url: 'mailto:support@interledger.app',
                        type: 'link',
                        children: [
                          { type: 'span', value: 'support@interledger.app' }
                        ]
                      },
                      { type: 'span', value: '.' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      { type: 'span', value: 'Please see ' },
                      {
                        url: 'https://www.gpo.gov/fdsys/pkg/USCODE-2011-title17/pdf/USCODE-2011-title17-chap5-sec512.pdf',
                        type: 'link',
                        children: [
                          { type: 'span', value: '17 U.S.C. § 512(c)(3)' }
                        ]
                      },
                      {
                        type: 'span',
                        value:
                          ' for the requirements of a proper notification. Also, please note that if you knowingly misrepresent that any activity or material on our Services is infringing, you may be liable to Corporation for certain costs and damages.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      { type: 'span', value: '11. Third-Party Content' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'We may provide information about third-party products, services, activities or events, or we may allow third parties to make their content and information available on or through the Services (collectively, “'
                      },
                      {
                        type: 'span',
                        marks: ['emphasis'],
                        value: 'Third-Party Content'
                      },
                      {
                        type: 'span',
                        value:
                          '”). We provide Third-Party Content as a service to those interested in such content. Your dealings or correspondence with third parties and your use of or interaction with any Third-Party Content are solely between you and the third party. Corporation does not control or endorse, and makes no representations or warranties regarding, any Third-Party Content, and your access to and use of such Third-Party Content is at your own risk.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [{ type: 'span', value: '12. Indemnification' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'To the fullest extent permitted by applicable law, you will indemnify, defend and hold harmless Corporation, and our past, present and future employees, officers, directors, contractors, consultants, equity holders, suppliers, vendors, service providers, parent companies, subsidiaries, affiliates, agents, representatives, predecessors, successors and assigns (individually and collectively, the “'
                      },
                      {
                        type: 'span',
                        marks: ['emphasis'],
                        value: 'Corporation Parties'
                      },
                      {
                        type: 'span',
                        value:
                          '”) from and against any losses, liabilities, claims, demands, actions, damages, expenses or costs (“'
                      },
                      { type: 'span', marks: ['emphasis'], value: 'Claims' },
                      {
                        type: 'span',
                        value:
                          "”) arising out of or related to (a) your access to or use of the Services; (b) your Associated Data or Feedback; (c) your violation of these Terms; (d) your violation, misappropriation or infringement of any rights of another (including intellectual property rights or privacy rights); or (e) your conduct in connection with the Services. You agree to promptly notify Corporation Parties of any third-party Claims, cooperate with Corporation Parties in defending such Claims and pay all fees, costs and expenses associated with defending such Claims (including attorneys' fees). You also agree that the Corporation Parties will have control of the defense or settlement, at Corporation's sole option, of any third-party Claims. This indemnity is in addition to, and not in lieu of, any other indemnities set forth in a written agreement between you and Corporation or the other Corporation Parties."
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [{ type: 'span', value: '13. Disclaimers' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Your use of our Services is at your sole risk. Except as otherwise provided in a writing by us, our Services and any content therein are provided “as is” and “as available” without warranties of any kind, either express or implied, including implied warranties of merchantability, fitness for a particular purpose, title, and non-infringement. In addition, Corporation does not represent or warrant that our Services are accurate, complete, reliable, current or error-free. While Corporation attempts to make your use of our Services and any content therein safe, we cannot and do not represent or warrant that our Services or servers are free of malware, viruses or other harmful components. You assume the entire risk as to the quality and performance of the Services.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Some jurisdictions do not allow the exclusion of certain warranties or disclaimer of implied terms in contracts with consumers, so some or all of the exclusions of warranties and disclaimers in this Section 13 may not apply to you.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      { type: 'span', value: '14. Limitation of Liability' }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'To the fullest extent permitted by applicable law, Corporation and the other Corporation Parties will not be liable to you under any theory of liability—whether based in contract, tort, negligence, strict liability, warranty, or otherwise—for any indirect, consequential, exemplary, incidental, punitive or special damages or lost profits, even if Corporation or the other Corporation Parties have been advised of the possibility of such damages.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The total liability of Corporation and the other Corporation Parties for any claim arising out of or relating to these Terms or our Services, regardless of the form of the action, is limited to the amount paid to the Corporation by you to use our Services.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The limitations set forth in this Section 14 will not limit or exclude liability for the gross negligence, fraud or intentional misconduct of Corporation or the other Corporation Parties or for any other matters in which liability cannot be excluded or limited under applicable law. Additionally, some jurisdictions do not allow the exclusion or limitation of incidental or consequential damages, so the above limitations or exclusions may not apply to you.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [{ type: 'span', value: '15. Release' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'To the fullest extent permitted by applicable law, you release Corporation and the other Corporation Parties from responsibility, liability, claims, demands and/or damages (actual and consequential) of every kind and nature, known and unknown (including claims of negligence), arising out of or related to (a) disputes between you and other users of the Services; (b) disputes between you and third parties that view, access, use, host or otherwise interact with your wallet or Associated Data; and (c) the acts or omissions of third parties. You waive any statute or common law principles that would otherwise limit the coverage of this release to include only those claims which you may know or suspect to exist in your favor at the time of agreeing to this release.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      {
                        type: 'span',
                        value: '16. Transfer and Processing Data'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'In order for us to provide our Services, you agree that we may process, transfer and store information about you in the United States and other countries, where you may not have the same rights and protections as you do under local law.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      {
                        type: 'span',
                        value: '17. Dispute Resolution; Binding Arbitration'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Please read the following section carefully because it requires you to arbitrate certain disputes and claims with Corporation and limits the manner in which you can seek relief from us, unless you opt out of arbitration by following the instructions set forth below. No class or representative actions or arbitrations are allowed under this arbitration provision. In addition, arbitration precludes you from suing in court or having a jury trial.'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'No Representative Actions. You and Corporation agree that any dispute arising out of or related to these Terms or our Services is personal to you and Corporation and that any dispute will be resolved solely through individual action, and will not be brought as a class arbitration, class action or any other type of representative proceeding.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Arbitration of Disputes. Except for (i) small claims disputes in which you or Corporation seeks to bring an individual action in small claims court located in the county of your billing address or (ii) disputes in which you or Corporation seeks injunctive or other equitable relief (x) to enforce this dispute resolution clause or (y) for the alleged infringement or misappropriation of intellectual property, including copyrights, trademarks, trade names, logos, trade secrets or patents, you and Corporation waive your rights to a jury trial and to have any other dispute arising out of or related to these Terms or our Services, including claims related to privacy and data security, (collectively, “'
                              },
                              {
                                type: 'span',
                                marks: ['emphasis'],
                                value: 'Disputes'
                              },
                              {
                                type: 'span',
                                value:
                                  '”) resolved in court. Instead, for any Dispute that you have against Corporation you agree to first contact Corporation and attempt to resolve the claim informally by sending a written notice of your claim (“'
                              },
                              {
                                type: 'span',
                                marks: ['emphasis'],
                                value: 'Notice'
                              },
                              {
                                type: 'span',
                                value: '”) to Corporation by email at '
                              },
                              {
                                url: 'mailto:support@fynbos.app',
                                type: 'link',
                                children: [
                                  {
                                    type: 'span',
                                    value: 'support@interledger.app'
                                  }
                                ]
                              },
                              {
                                type: 'span',
                                value:
                                  ' or by certified mail addressed to Interledger Inc, 447 Broadway, 2nd Floor Suite #2233, New York, 10013. The Notice must (I) include your name, residence address, email address, and telephone number; (II) describe the nature and basis of the Dispute; and (III) set forth the specific relief sought. Our notice to you will be similar in form to that described above. If you and Corporation cannot reach an agreement to resolve the Dispute within thirty (30) days after such Notice is received, then either party may submit the Dispute to binding arbitration administered by JAMS or, under the limited circumstances set forth above, in court. All Disputes submitted to JAMS will be resolved through confidential, binding arbitration before one arbitrator. Arbitration proceedings will be held in, Delaware unless you are a consumer, in which case you may elect to hold the arbitration in your county of residence. For purposes of this Section 17, a “'
                              },
                              {
                                type: 'span',
                                marks: ['emphasis'],
                                value: 'consumer'
                              },
                              {
                                type: 'span',
                                value:
                                  '” means a person using the Services for personal, family or household purposes. You and Corporation agree that Disputes will be held in accordance with the JAMS Streamlined Arbitration Rules and Procedures (“'
                              },
                              {
                                type: 'span',
                                marks: ['emphasis'],
                                value: 'JAMS Rules'
                              },
                              {
                                type: 'span',
                                value:
                                  '”). The most recent version of the JAMS Rules are available on the '
                              },
                              {
                                url: 'https://www.jamsadr.com/rules-streamlined-arbitration/',
                                type: 'link',
                                children: [
                                  { type: 'span', value: 'JAMS website' }
                                ]
                              },
                              {
                                type: 'span',
                                value:
                                  ' and are hereby incorporated by reference. You either acknowledge and agree that you have read and understand the JAMS Rules or waive your opportunity to read the JAMS Rules and waive any claim that the JAMS Rules are unfair or should not apply for any reason.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You and Corporation agree that these Terms affect interstate commerce and that the enforceability of this Section 17 will be substantively and procedurally governed by the Federal Arbitration Act, 9 U.S.C. § 1, '
                              },
                              {
                                type: 'span',
                                marks: ['emphasis'],
                                value: 'et seq'
                              },
                              { type: 'span', value: '. (the “' },
                              {
                                type: 'span',
                                marks: ['emphasis'],
                                value: 'FAA'
                              },
                              {
                                type: 'span',
                                value:
                                  '”), to the maximum extent permitted by applicable law. As limited by the FAA, these Terms and the JAMS Rules, the arbitrator will have exclusive authority to make all procedural and substantive decisions regarding any Dispute and to grant any remedy that would otherwise be available in court. The arbitrator may conduct only an individual arbitration and may not consolidate more than one individual’s claims, preside over any type of class or representative proceeding or preside over any proceeding involving more than one individual.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The arbitration will allow for the discovery or exchange of non-privileged information relevant to the Dispute.The arbitrator, Corporation, and you will maintain the confidentiality of any arbitration proceedings, judgments and awards, including information gathered, prepared and presented for purposes of the arbitration or related to the Dispute(s) therein. The arbitrator will have the authority to make appropriate rulings to safeguard confidentiality, unless the law provides to the contrary. The duty of confidentiality does not apply to the extent that disclosure is necessary to prepare for or conduct the arbitration hearing on the merits, in connection with a court application for a preliminary remedy or in connection with a judicial challenge to an arbitration award or its enforcement, or to the extent that disclosure is otherwise required by law or judicial decision.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You and Corporation agree that for any arbitration you initiate, you will pay the filing fee (up to a maximum of $250 if you are a consumer), and Corporation will pay the remaining JAMS fees and costs. For any arbitration initiated by Corporation, Corporation will pay all JAMS fees and costs. You and Corporation agree that the state or federal courts of the State of Delaware and the United States have exclusive jurisdiction over any appeals and the enforcement of an arbitration award.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Any Dispute must be filed within one year after the relevant claim arose; otherwise, the Dispute is permanently barred, which means that you and Corporation will not have the right to assert the claim.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You have the right to opt out of binding arbitration within 30 days of the date you first accepted the terms of this Section 17 by sending a written notice to the Corporation by email at '
                              },
                              {
                                url: 'mailto:support@interledger.app',
                                type: 'link',
                                children: [
                                  {
                                    type: 'span',
                                    value: 'support@interledger.app'
                                  }
                                ]
                              },
                              {
                                type: 'span',
                                value:
                                  ' or by certified mail addressed to Interledger Inc, 447 Broadway, 2nd Floor Suite #2233, New York, 10013. In order to be effective, the opt-out notice must include your full name and address and clearly indicate your intent to opt out of binding arbitration. By opting out of binding arbitration, you are agreeing to resolve Disputes in accordance with Section 17.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'If any portion of this Section 17 is found to be unenforceable or unlawful for any reason, (i) the unenforceable or unlawful provision shall be severed from these Terms; (ii) severance of the unenforceable or unlawful provision shall have no impact whatsoever on the remainder of this Section 17 or the parties’ ability to compel arbitration of any remaining claims on an individual basis pursuant to this Section 17; and (iii) to the extent that any claims must therefore proceed on a class, collective, consolidated, or representative basis, such claims must be litigated in a civil court of competent jurisdiction and not in arbitration, and the parties agree that litigation of those claims shall be stayed pending the outcome of any individual claims in arbitration. Further, if any part of this Section 17 is found to prohibit an individual claim seeking public injunctive relief, that provision will have no effect to the extent such relief is allowed to be sought out of arbitration, and the remainder of this Section 17 will be enforceable.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      { type: 'span', value: '18. Governing Law and Venue' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Any dispute arising from these Terms and your use of the Services will be governed by and construed and enforced in accordance with the laws of Delaware, except to the extent preempted by U.S. federal law, without regard to conflict of law rules or principles (whether of Delaware or any other jurisdiction) that would cause the application of the laws of any other jurisdiction. Any dispute between the parties that is not subject to arbitration or cannot be heard in small claims court will be resolved in the state or federal courts of Delaware and the United States, respectively.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      {
                        type: 'span',
                        value: '19. Modifying and Terminating our Services'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'We reserve the right to modify our Services or to suspend or stop providing all or portions of our Services at any time. You also have the right to stop using our Services at any time. We are not responsible for any loss or harm related to your inability to access or use our Services. If we discontinue providing all or portions of the Services, we will, where reasonably possible, give you advance notice.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [
                      {
                        type: 'span',
                        value: '20. Additional Terms and Amendments'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'We may supply different or additional terms in relation to some of our Services, and those different or additional terms become part of your agreement with us if you use those Services. If there is a conflict between these Terms and the additional terms, the additional terms will control for that conflict.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'We may make changes to these Terms from time to time. If we make changes, we will provide you with notice of such changes, such as by sending an email, providing a notice through our Services or updating the date at the top of these Terms. Unless we say otherwise in our notice, the amended Terms will be effective immediately, and your continued use of our Services after we provide such notice will confirm your acceptance of the changes. If you do not agree to the amended Terms, you must stop using our Services.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The current version of the license terms for wallet addresses may be found at '
                              },
                              {
                                url: 'https://interledger.app/legal',
                                type: 'link',
                                children: [
                                  {
                                    type: 'span',
                                    value: 'https://interledger.app/legal'
                                  }
                                ]
                              },
                              {
                                type: 'span',
                                value:
                                  ' and are hereby incorporated by reference. You acknowledge and agree that you have read and understand the License Terms and agree to be bound by its terms.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [{ type: 'span', value: '21. Severability' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'If any provision or part of a provision of these Terms is unlawful, void or unenforceable, that provision or part of the provision is deemed severable from these Terms and does not affect the validity and enforceability of any remaining provisions.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 1,
                    children: [{ type: 'span', value: '22. Miscellaneous' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'The failure of Corporation to exercise or enforce any right or provision of these Terms will not operate as a waiver of such right or provision. These Terms reflect the entire agreement between the parties relating to the subject matter hereof and supersede all prior agreements, representations, statements and understandings of the parties, whether express or implied. The section titles in these Terms are for convenience only and have no legal or contractual effect. Use of the word “including” will be interpreted to mean “including without limitation.” You may not assign your rights and obligations under these Terms without our express written consent. Our failure to exercise or enforce any right or provision of these Terms will not operate as a waiver of such right or provision. We will not be liable for any delay or failure to perform any obligation under these Terms where the delay or failure results from any cause beyond our reasonable control. Your access to or use of the Services does not create any form of partnership, joint venture or any other similar relationship between you and us. Except as otherwise provided herein, these Terms are intended solely for the benefit of the parties and are not intended to confer third-party beneficiary rights upon any other person or entity. You agree that communications and transactions between us may be conducted electronically.'
                      }
                    ]
                  }
                ]
              }
            },
            __typename: 'LegalPageModelBodyField'
          },
          id: '125234304',
          title: 'Terms of Service',
          _publishedAt: '2024-07-15T16:07:23+02:00',
          _seoMetaTags: [
            {
              tag: 'title',
              attributes: null,
              content: 'Terms of Service',
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:title', content: 'Terms of Service' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                name: 'twitter:title',
                content: 'Terms of Service'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'twitter:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:locale', content: 'en' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:type', content: 'article' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:site_name', content: 'Interledger' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                property: 'article:modified_time',
                content: '2025-04-06T16:23:03Z'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'article:publisher', content: '' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                name: 'twitter:card',
                content: 'summary_large_image'
              },
              content: null,
              __typename: 'Tag'
            }
          ],
          __typename: 'LegalPageRecord'
        } as Query['legalPage'],
        footer: {
          __typename: 'FooterRecord',
          id: '121270040',
          logo: {
            id: 'dffIomSsTfCkd-b3vjChtA',
            url: 'https://www.datocms-assets.com/160242/1721916494-interledger_icon.svg',
            height: 101,
            width: 101,
            __typename: 'FileField'
          },
          column1Title: 'Menu',
          column1: [
            {
              id: '125075096',
              displayText: 'Contact',
              url: 'https://interledger.app/contact',
              __typename: 'LinkRecord'
            }
          ],
          column2Title: 'Ecosystem',
          column2: [
            {
              id: '121270066',
              displayText: 'Interledger Foundation',
              url: 'https://interledger.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075093',
              displayText: 'Web monetization',
              url: 'https://webmonetization.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075094',
              displayText: 'Open Payments',
              url: 'https://openpayments.dev/',
              __typename: 'LinkRecord'
            }
          ],
          column3Title: 'Legal',
          column3: [
            {
              id: '121270067',
              displayText: 'Legal Agreements',
              url: 'https://interledger.app/legal',
              __typename: 'LinkRecord'
            }
          ],
          legalText: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          '© 2024 Interledger Inc and the Interledger Foundation.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'The Interledger name and logo are the property of the'
                      },
                      {
                        url: 'https://interledger.org',
                        type: 'link',
                        children: [
                          { type: 'span', value: ' Interledger Foundation' }
                        ]
                      },
                      {
                        type: 'span',
                        value:
                          '. The Interledger app is powered by Interledger on behalf of the Interledger Foundation as a service to the Interledger community. Interledger Inc is not a bank. Interledger provides a technology platform and all payments and banking services are provided by our partners who are appropriately licensed.'
                      }
                    ]
                  }
                ]
              }
            },
            __typename: 'FooterModelLegalTextField'
          },
          socialIcons: [
            {
              id: '124003009',
              url: 'https://x.com/Interledger',
              icon: {
                id: 'If7vlze6QAeqjixiyS922g',
                url: 'https://www.datocms-assets.com/160242/1710406464-x-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089438',
              url: 'https://www.linkedin.com/company/interledger-foundation/',
              icon: {
                id: '52090585',
                url: 'https://www.datocms-assets.com/160242/1685976901-icon-linkedin.svg',
                height: 16,
                width: 16,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089519',
              url: 'https://www.youtube.com/@InterledgerFoundation',
              icon: {
                id: '52090589',
                url: 'https://www.datocms-assets.com/160242/1685977020-icon-youtube.svg',
                height: 13,
                width: 19,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'PCadAp0bSRW4_0v6BTeiKg',
              url: 'https://www.instagram.com/interledgerfoundation/',
              icon: {
                id: 'dfWXhnzmQT68juUEGsvCkw',
                url: 'https://www.datocms-assets.com/160242/1710227274-instagram-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'JFxCqHyWQCyJbpfrhMbeJQ',
              url: 'https://www.facebook.com/interledger',
              icon: {
                id: 'YhE1LNyHRSWKFCxYuStcYg',
                url: 'https://www.datocms-assets.com/160242/1710227315-facebook-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            }
          ]
        } as Query['footer']
      }
    case 'privacy-policy': {
      return {
        legalPage: {
          slug: 'privacy-policy',
          external: '',
          body: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Interledger ("we," "us," or "our") respects your privacy and is committed to protecting your personal data. This Privacy Policy ("Policy") will inform you about how we collect, use, share, and protect your personal data when you use our online digital wallet services (collectively, the "Services").'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'This Policy also describes your data protection rights and how you can exercise them.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Please read this Policy carefully to understand our policies and practices regarding your personal data.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 2,
                    children: [
                      { type: 'span', value: 'Personal Data We Collect' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'We collect and process the following categories of personal data from you when you use our Services:'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              { type: 'span', value: 'Full legal name' }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              { type: 'span', value: 'Physical address' }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [{ type: 'span', value: 'Email address' }]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [{ type: 'span', value: 'Mobile number' }]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [{ type: 'span', value: 'Date of birth' }]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              { type: 'span', value: 'Biometric data' }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              { type: 'span', value: 'Payment card details' }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              { type: 'span', value: 'Bank account details' }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value: 'Social media account details'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [{ type: 'span', value: 'IP Address' }]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              { type: 'span', value: 'Approximate location' }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value: 'Device and browser fingerprints'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value: 'How We Use Your Personal Data'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'We use your personal data for the following purposes:'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value: 'Register you as a user of our Services'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Contact you regarding your account or transactions'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Verify your legal identity to prevent fraud and comply with regulatory requirements, such as the Bank Secrecy Act and Patriot Act'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Comply with legal, regulatory, and contractual obligations'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Reduce the risk of, or prevent, fraudulent use of our Services'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Execute transactions with counterparties that need to verify your personal information'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value: 'No Sale of Personal Data'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Interledger will never sell your personal data or disclose it to a third-party unless required to do so to deliver the Services to you.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 2,
                    children: [
                      { type: 'span', value: 'Sharing Your Personal Data' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'We share your personal data with the following external processors to assist us in providing our Services:'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Fiant Inc - for facilitating money transfers between users and across different financial institutions in the USA'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Gatehub and Paywiser - for facilitating money transfers between users and across different financial institutions in the EU'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Xago  - for facilitating money transfers between users and across different financial institutions in South Africa'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Astra - for processing payment transactions and verifying payment card information'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Persona - for identity verification, fraud prevention, and compliance with regulatory requirements'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Grafana Cloud - for monitoring, analyzing, and visualizing data related to the performance of our Services'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Segment - for tracking user interactions and events to improve our Services'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Twilio - for communication purposes, including SMS and email notifications'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'These external processors are contractually obligated to process your personal data in accordance with applicable data protection laws and regulations.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 2,
                    children: [{ type: 'span', value: 'Data Security' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'We have implemented appropriate technical and organizational measures to ensure the security of your personal data. These measures are designed to protect your personal data from unauthorized access, use, disclosure, alteration, and destruction.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 2,
                    children: [{ type: 'span', value: 'Data Retention' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'We will retain your personal data for as long as necessary to fulfill the purposes for which it was collected, including for the purposes of satisfying any legal, regulatory, accounting, or reporting requirements.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 2,
                    children: [{ type: 'span', value: 'Your Rights' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      { type: 'span', value: 'You have the right to:' }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value: 'Access your personal data'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Request correction of any inaccurate personal data'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Request deletion of your personal data in certain circumstances'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Object to the processing of your personal data for direct marketing purposes'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Request restriction of processing your personal data in certain circumstances'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Data portability, allowing you to obtain and reuse your personal data for your own purposes'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 2,
                    children: [
                      { type: 'span', value: 'Rights for California Residents' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'If you are a California resident, you have the following additional rights under the California Consumer Privacy Act (CCPA):'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to know what personal information has been collected about you, the sources from which it was collected, and the purpose for collecting it.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to know whether your personal information is sold or disclosed and to whom.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to opt-out of the sale of your personal information.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to non-discrimination for exercising your CCPA rights. This means we will not deny you services, charge you different prices or provide a different level or quality of services, based solely on the exercise of your CCPA rights.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'To exercise any of these rights, please contact us using the contact details provided below. Please note that we may need to verify your identity before processing your request.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 2,
                    children: [
                      { type: 'span', value: 'Rights for EU Residents (GDPR)' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'If you are a resident of the European Union, you have the following rights under the General Data Protection Regulation (GDPR):'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to be informed about the collection and use of your personal data.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to access your personal data and supplementary information.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to rectify any inaccurate or incomplete personal data.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to have your personal data erased (the "right to be forgotten") under certain circumstances.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to restrict the processing of your personal data under certain circumstances.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to data portability, allowing you to obtain and reuse your personal data across different services.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to object to the processing of your personal data for direct marketing purposes or when processing is based on legitimate interests or the performance of a task in the public interest.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to not be subject to a decision based solely on automated processing, including profiling, which has legal or similarly significant effects.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'To exercise any of these rights, please contact us using the contact details provided below. Please note that we may need to verify your identity before processing your request.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 2,
                    children: [
                      {
                        type: 'span',
                        value: 'Rights for Brazilian Residents (LGPD)'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'If you are a resident of Brazil, you have the following rights under the Brazilian General Data Protection Law (LGPD):'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to confirm the existence of the processing of your personal data.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value: 'The right to access your personal data.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to correct incomplete, inaccurate, or outdated personal data.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to anonymize, block, or delete unnecessary or excessive personal data, or data that is not being processed in compliance with the LGPD.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to the portability of your personal data to another service or product provider, by means of an express request.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to delete your personal data processed with your consent, except in cases where the law requires or allows the data to be retained.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to information about public and private entities with which we have shared your personal data.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to refuse consent and to be informed of the consequences of such refusal.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The right to revoke consent at any time.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'To exercise any of these rights, please contact us using the contact details provided below. Please note that we may need to verify your identity before processing your request.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 2,
                    children: [
                      { type: 'span', value: 'Changes to This Policy' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'We may update this Policy from time to time. We will notify you of any changes by posting the new Policy on our website. It is your responsibility to review this Policy periodically for any changes.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 2,
                    children: [{ type: 'span', value: 'Contact Us' }]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'If you have any questions or concerns about this Policy or our data practices, please contact us at '
                      },
                      {
                        url: 'mailto:support@interledger.app',
                        type: 'link',
                        children: [
                          { type: 'span', value: 'support@interledger.app' }
                        ]
                      },
                      { type: 'span', value: '.' }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'By using our Services, you agree to the terms of this Policy.'
                      }
                    ]
                  }
                ]
              }
            },
            __typename: 'LegalPageModelBodyField'
          },
          id: '125282647',
          title: 'Privacy Policy',
          _publishedAt: '2024-07-15T16:07:24+02:00',
          _seoMetaTags: [
            {
              tag: 'title',
              attributes: null,
              content: 'Privacy Policy',
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:title', content: 'Privacy Policy' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'twitter:title', content: 'Privacy Policy' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'twitter:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:locale', content: 'en' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:type', content: 'article' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:site_name', content: 'Interledger' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                property: 'article:modified_time',
                content: '2025-04-06T16:23:03Z'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'article:publisher', content: '' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                name: 'twitter:card',
                content: 'summary_large_image'
              },
              content: null,
              __typename: 'Tag'
            }
          ],
          __typename: 'LegalPageRecord'
        } as Query['legalPage'],
        footer: {
          __typename: 'FooterRecord',
          id: '121270040',
          logo: {
            id: 'dffIomSsTfCkd-b3vjChtA',
            url: 'https://www.datocms-assets.com/160242/1721916494-interledger_icon.svg',
            height: 101,
            width: 101,
            __typename: 'FileField'
          },
          column1Title: 'Menu',
          column1: [
            {
              id: '125075096',
              displayText: 'Contact',
              url: 'https://interledger.app/contact',
              __typename: 'LinkRecord'
            }
          ],
          column2Title: 'Ecosystem',
          column2: [
            {
              id: '121270066',
              displayText: 'Interledger Foundation',
              url: 'https://interledger.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075093',
              displayText: 'Web monetization',
              url: 'https://webmonetization.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075094',
              displayText: 'Open Payments',
              url: 'https://openpayments.dev/',
              __typename: 'LinkRecord'
            }
          ],
          column3Title: 'Legal',
          column3: [
            {
              id: '121270067',
              displayText: 'Legal Agreements',
              url: 'https://interledger.app/legal',
              __typename: 'LinkRecord'
            }
          ],
          legalText: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          '© 2024 Interledger Inc and the Interledger Foundation.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'The Interledger name and logo are the property of the'
                      },
                      {
                        url: 'https://interledger.org',
                        type: 'link',
                        children: [
                          { type: 'span', value: ' Interledger Foundation' }
                        ]
                      },
                      {
                        type: 'span',
                        value:
                          '. The Interledger app is powered by Interledger on behalf of the Interledger Foundation as a service to the Interledger community. Interledger Inc is not a bank. Interledger provides a technology platform and all payments and banking services are provided by our partners who are appropriately licensed.'
                      }
                    ]
                  }
                ]
              }
            },
            __typename: 'FooterModelLegalTextField'
          },
          socialIcons: [
            {
              id: '124003009',
              url: 'https://x.com/Interledger',
              icon: {
                id: 'If7vlze6QAeqjixiyS922g',
                url: 'https://www.datocms-assets.com/160242/1710406464-x-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089438',
              url: 'https://www.linkedin.com/company/interledger-foundation/',
              icon: {
                id: '52090585',
                url: 'https://www.datocms-assets.com/160242/1685976901-icon-linkedin.svg',
                height: 16,
                width: 16,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089519',
              url: 'https://www.youtube.com/@InterledgerFoundation',
              icon: {
                id: '52090589',
                url: 'https://www.datocms-assets.com/160242/1685977020-icon-youtube.svg',
                height: 13,
                width: 19,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'PCadAp0bSRW4_0v6BTeiKg',
              url: 'https://www.instagram.com/interledgerfoundation/',
              icon: {
                id: 'dfWXhnzmQT68juUEGsvCkw',
                url: 'https://www.datocms-assets.com/160242/1710227274-instagram-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'JFxCqHyWQCyJbpfrhMbeJQ',
              url: 'https://www.facebook.com/interledger',
              icon: {
                id: 'YhE1LNyHRSWKFCxYuStcYg',
                url: 'https://www.datocms-assets.com/160242/1710227315-facebook-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            }
          ]
        } as Query['footer']
      }
    }
    case 'wallet-license': {
      return {
        legalPage: {
          slug: 'wallet-license',
          external: '',
          body: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'If you have registered and activated an Interledger wallet (as defined herein), Interledger Foundation ("Interledger") grants you a license to use the wallet addresses associated with that wallet in accordance with the following terms (the "Terms"). Please note that your license solely applies to the wallet address URL as specified by you at the time the wallet address was activated and does not confer any rights to use any future wallet addresses Interledger may activate nor any third party wallet addresses or URLs which are not owned by Interledger.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Here’s a summary of how, in general terms, you may and may not use your wallet address:'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'bulleted',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may use your wallet address on your social media pages and in your social media handles.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may print your wallet address on merchandise for commercial sale.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may create a brand around your wallet address.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may display your wallet address in images, videos, or animations.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may use your wallet to advertise your business.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may use your wallet in a political campaign.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may auction your wallet address for charity.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may incorporate your wallet address into derivative works.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may NOT transfer your wallet address to a different wallet.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may NOT change or modify the URL comprising your wallet address.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'You may NOT use someone else’s wallet address.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Your right to use a wallet address terminates if you violate these Terms.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Please read the following Terms carefully as they set out the above as well as other contractual terms which will be binding on you:'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Interledger hereby grants to each wallet address holder ("Holder") a license to the wallet address activated by such Holder, upon the following terms and conditions and the other provisions of these Terms:'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The license is perpetual, non-exclusive, royalty-free, and world-wide.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'For any Holder, the license to use a wallet address or any derivative work from a wallet address commences when such Holder activates the wallet address and terminates automatically when such Holder violates any of these Terms or at the express discretion of Interledger.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Each Holder may copy, display, distribute, and create derivative works of the wallet address for any commercial or non-commercial use (the “Personal Use Right”). The Holder agrees that the Personal Use Right is to use a whole and entire address, and not any part thereof or form of display which would obscure or diminish any individual part of a wallet address such that the wallet address resembles a different wallet.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Each Holder may use his or her wallet or wallet address without attribution to Interledger.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The Holder may not, transfer, or assign any of its right, title, and interest in and to its license to use a given wallet address.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The Holder may not delegate, sublicense, temporarily assign, mortgage, charge, or pledge the Holder’s Personal Use Right.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Except as otherwise provided herein, the Personal Use Right does not grant any rights to use the word "Interledger" or the Interledger logo or any trademark of the Interledger Foundation in connection with the use of any wallet.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Interledger retains exclusive interest in and ownership of its intellectual property, including without limitation all patents, copyrights, trademarks (together with the goodwill symbolized thereby), trade secrets, know-how, and other confidential or proprietary information, and other intellectual property rights (collectively "Intellectual Property Rights"). Each Holder acknowledges and agrees that Interledger\'s Intellectual Property Rights subsist in, and Interledger retains all Intellectual Property Rights in, each and any wallet address, part thereof or combination thereof.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          "Except for the grant of the Personal Use Right in accordance with these Terms, nothing in these Terms functions to assign or transfer, nor creates any right in favor of any person to use, any of Interledger's Intellectual Property Rights."
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Each Holder will not, and agrees not to cause or allow any other person to:'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'use the wallet address in any way that exceeds the scope of the Holder’s license to the wallet address;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'use the wallet or wallet address with material that violates any third-party rights, or otherwise take any action in connection with the wallet that infringes the intellectual property or other rights of any person or entity;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'register or apply to register, object to the registration of, seek to cancel, or otherwise claim or assert rights in, a trademark, design mark, service mark, sound mark, or tradename (“Mark”), that uses any Third Party wallet address; claim or enforce ownership rights in any Third Party wallet address; or make any attempt to prevent any third party from using any Third Party wallet address, regardless of the degree of similarity between the Holder’s Mark and the Third Party wallet address or any other legal rights the Holder may have in relation to the use of a Third Party wallet address, whether at common law, equity, statute, or otherwise. “Third Party wallet address” means any wallet address, part thereof or combination thereof, in whole or in part, which a Holder does not have a license to use. This Section shall survive any termination of these Terms;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'use the wallet address in a manner that is tortious, defamatory, in association with any unlawful goods or service, or in any way that violates any applicable laws, rules, or regulations;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'remove, obscure or alter any proprietary notices associated with the wallet address, or give any express or implied misrepresentation that the Holder or another third party are Interledger or the holder of the copyright or other applicable intellectual property rights in any wallet address;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'use the wallet address other than for the benefit of the Holder;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'use or exploit the wallet address in any manner other than as expressly permitted in these Terms;'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'modify the wallet address or any part thereof.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'For the avoidance of doubt, in order to use wallet addresses for a purpose not authorized by these Terms, the Holder must first (1) obtain a license directly from Interledger; and (2) secure additional permissions as necessary. Interledger shall be under no obligation to grant or negotiate or offer such additional license and may either grant or withhold such license in its sole and absolute discretion.'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'numbered',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  "Indemnification Obligations. Without limiting the obligations in these Terms, each Holder agrees to indemnify, hold harmless, compensate and reimburse Interledger and its respective subsidiaries, affiliates, officers, agents, employees, partners, and licensors from or for any claim, demand, loss, or damages, including reasonable attorneys' fees, arising out of or related to Holder's use of the wallet address or Holder's violation of these Terms."
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  "Injunctive Relief. Notwithstanding anything else in these Terms, Holder hereby agrees that, in the event of Holder's or any third partys unauthorized access to, or use of, the wallet address in violation of these Terms, Interledger shall be entitled to apply for injunctive remedies (or an equivalent type of urgent legal relief) in any jurisdiction, without providing notice or opportunity to cure."
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Amendments. Interledger may make changes to these Terms from time to time. If Interledger makes changes, we may provide you with notice of such changes, such as by sending an email, providing a notice through our Services or updating the date at the top of these Terms. Unless we state otherwise, the amended terms will be effective immediately, and your continued use of our Services after we provide such notice will confirm your acceptance of the changes.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Part of Terms of Service. These License Terms form part of the interledger.app Terms of Service which may be found at '
                              },
                              {
                                url: 'https://interledger.app/legal',
                                type: 'link',
                                children: [
                                  {
                                    type: 'span',
                                    value: 'interledger.app/legal'
                                  }
                                ]
                              },
                              {
                                type: 'span',
                                value:
                                  ' and the terms of the interledger.app Terms of Service are expressly incorporated herein by reference.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  }
                ]
              }
            },
            __typename: 'LegalPageModelBodyField'
          },
          id: '125282648',
          title: 'Wallet License',
          _publishedAt: '2024-07-15T16:07:24+02:00',
          _seoMetaTags: [
            {
              tag: 'title',
              attributes: null,
              content: 'Wallet License',
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:title', content: 'Wallet License' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'twitter:title', content: 'Wallet License' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'twitter:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:locale', content: 'en' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:type', content: 'article' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:site_name', content: 'Interledger' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                property: 'article:modified_time',
                content: '2025-04-06T16:23:03Z'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'article:publisher', content: '' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                name: 'twitter:card',
                content: 'summary_large_image'
              },
              content: null,
              __typename: 'Tag'
            }
          ],
          __typename: 'LegalPageRecord'
        } as Query['legalPage'],
        footer: {
          __typename: 'FooterRecord',
          id: '121270040',
          logo: {
            id: 'dffIomSsTfCkd-b3vjChtA',
            url: 'https://www.datocms-assets.com/160242/1721916494-interledger_icon.svg',
            height: 101,
            width: 101,
            __typename: 'FileField'
          },
          column1Title: 'Menu',
          column1: [
            {
              id: '125075096',
              displayText: 'Contact',
              url: 'https://interledger.app/contact',
              __typename: 'LinkRecord'
            }
          ],
          column2Title: 'Ecosystem',
          column2: [
            {
              id: '121270066',
              displayText: 'Interledger Foundation',
              url: 'https://interledger.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075093',
              displayText: 'Web monetization',
              url: 'https://webmonetization.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075094',
              displayText: 'Open Payments',
              url: 'https://openpayments.dev/',
              __typename: 'LinkRecord'
            }
          ],
          column3Title: 'Legal',
          column3: [
            {
              id: '121270067',
              displayText: 'Legal Agreements',
              url: 'https://interledger.app/legal',
              __typename: 'LinkRecord'
            }
          ],
          legalText: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          '© 2024 Interledger Inc and the Interledger Foundation.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'The Interledger name and logo are the property of the'
                      },
                      {
                        url: 'https://interledger.org',
                        type: 'link',
                        children: [
                          { type: 'span', value: ' Interledger Foundation' }
                        ]
                      },
                      {
                        type: 'span',
                        value:
                          '. The Interledger app is powered by Interledger on behalf of the Interledger Foundation as a service to the Interledger community. Interledger Inc is not a bank. Interledger provides a technology platform and all payments and banking services are provided by our partners who are appropriately licensed.'
                      }
                    ]
                  }
                ]
              }
            },
            __typename: 'FooterModelLegalTextField'
          },
          socialIcons: [
            {
              id: '124003009',
              url: 'https://x.com/Interledger',
              icon: {
                id: 'If7vlze6QAeqjixiyS922g',
                url: 'https://www.datocms-assets.com/160242/1710406464-x-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089438',
              url: 'https://www.linkedin.com/company/interledger-foundation/',
              icon: {
                id: '52090585',
                url: 'https://www.datocms-assets.com/160242/1685976901-icon-linkedin.svg',
                height: 16,
                width: 16,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089519',
              url: 'https://www.youtube.com/@InterledgerFoundation',
              icon: {
                id: '52090589',
                url: 'https://www.datocms-assets.com/160242/1685977020-icon-youtube.svg',
                height: 13,
                width: 19,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'PCadAp0bSRW4_0v6BTeiKg',
              url: 'https://www.instagram.com/interledgerfoundation/',
              icon: {
                id: 'dfWXhnzmQT68juUEGsvCkw',
                url: 'https://www.datocms-assets.com/160242/1710227274-instagram-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'JFxCqHyWQCyJbpfrhMbeJQ',
              url: 'https://www.facebook.com/interledger',
              icon: {
                id: 'YhE1LNyHRSWKFCxYuStcYg',
                url: 'https://www.datocms-assets.com/160242/1710227315-facebook-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            }
          ]
        } as Query['footer']
      }
    }

    case 'accessibility-statement': {
      return {
        legalPage: {
          slug: 'accessibility-statement',
          external: '',
          body: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Interledger Foundation is dedicated to providing an accessible and inclusive experience for all of our services, for all users, including those who use assistive technologies such as screen reading software, screen enlargement software, and alternative keyboard input devices. We firmly believe that everyone should have equal access to our services and are committed to continuously improving the accessibility of our platform.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'In order to achieve this goal, Interledger Foundation actively strives to implement the Web Content Accessibility Guidelines (WCAG) version 2.0/2.1 Level AA standard across all content and elements of our website. We regularly review and update our website to ensure it meets the evolving accessibility standards and best practices. Additionally, Interledger Foundation works closely with an assistive technology vendor to carry out periodic testing of our website using various assistive technologies, which helps us ensure full compliance with WCAG 2.0/2.1 Level AA.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'We appreciate feedback from our users to help us improve the accessibility of our website. If you use assistive technologies and encounter any difficulties when accessing our website, please do not hesitate to contact us at '
                      },
                      {
                        url: 'mailto:support@interledger.app',
                        type: 'link',
                        children: [
                          { type: 'span', value: 'support@interledger.app' }
                        ]
                      },
                      {
                        type: 'span',
                        value:
                          '. Our team is dedicated to promptly addressing your concerns and working towards providing a seamless experience for all users.'
                      }
                    ]
                  }
                ]
              }
            },
            __typename: 'LegalPageModelBodyField'
          },
          id: '125282649',
          title: 'Accessibility Statement',
          _publishedAt: '2024-07-15T16:07:24+02:00',
          _seoMetaTags: [
            {
              tag: 'title',
              attributes: null,
              content: 'Accessibility Statement',
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                property: 'og:title',
                content: 'Accessibility Statement'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                name: 'twitter:title',
                content: 'Accessibility Statement'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'twitter:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:locale', content: 'en' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:type', content: 'article' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:site_name', content: 'Interledger' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                property: 'article:modified_time',
                content: '2025-04-06T16:23:03Z'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'article:publisher', content: '' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                name: 'twitter:card',
                content: 'summary_large_image'
              },
              content: null,
              __typename: 'Tag'
            }
          ],
          __typename: 'LegalPageRecord'
        } as Query['legalPage'],
        footer: {
          __typename: 'FooterRecord',
          id: '121270040',
          logo: {
            id: 'dffIomSsTfCkd-b3vjChtA',
            url: 'https://www.datocms-assets.com/160242/1721916494-interledger_icon.svg',
            height: 101,
            width: 101,
            __typename: 'FileField'
          },
          column1Title: 'Menu',
          column1: [
            {
              id: '125075096',
              displayText: 'Contact',
              url: 'https://interledger.app/contact',
              __typename: 'LinkRecord'
            }
          ],
          column2Title: 'Ecosystem',
          column2: [
            {
              id: '121270066',
              displayText: 'Interledger Foundation',
              url: 'https://interledger.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075093',
              displayText: 'Web monetization',
              url: 'https://webmonetization.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075094',
              displayText: 'Open Payments',
              url: 'https://openpayments.dev/',
              __typename: 'LinkRecord'
            }
          ],
          column3Title: 'Legal',
          column3: [
            {
              id: '121270067',
              displayText: 'Legal Agreements',
              url: 'https://interledger.app/legal',
              __typename: 'LinkRecord'
            }
          ],
          legalText: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          '© 2024 Interledger Inc and the Interledger Foundation.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'The Interledger name and logo are the property of the'
                      },
                      {
                        url: 'https://interledger.org',
                        type: 'link',
                        children: [
                          { type: 'span', value: ' Interledger Foundation' }
                        ]
                      },
                      {
                        type: 'span',
                        value:
                          '. The Interledger app is powered by Interledger on behalf of the Interledger Foundation as a service to the Interledger community. Interledger Inc is not a bank. Interledger provides a technology platform and all payments and banking services are provided by our partners who are appropriately licensed.'
                      }
                    ]
                  }
                ]
              }
            },
            __typename: 'FooterModelLegalTextField'
          },
          socialIcons: [
            {
              id: '124003009',
              url: 'https://x.com/Interledger',
              icon: {
                id: 'If7vlze6QAeqjixiyS922g',
                url: 'https://www.datocms-assets.com/160242/1710406464-x-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089438',
              url: 'https://www.linkedin.com/company/interledger-foundation/',
              icon: {
                id: '52090585',
                url: 'https://www.datocms-assets.com/160242/1685976901-icon-linkedin.svg',
                height: 16,
                width: 16,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089519',
              url: 'https://www.youtube.com/@InterledgerFoundation',
              icon: {
                id: '52090589',
                url: 'https://www.datocms-assets.com/160242/1685977020-icon-youtube.svg',
                height: 13,
                width: 19,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'PCadAp0bSRW4_0v6BTeiKg',
              url: 'https://www.instagram.com/interledgerfoundation/',
              icon: {
                id: 'dfWXhnzmQT68juUEGsvCkw',
                url: 'https://www.datocms-assets.com/160242/1710227274-instagram-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'JFxCqHyWQCyJbpfrhMbeJQ',
              url: 'https://www.facebook.com/interledger',
              icon: {
                id: 'YhE1LNyHRSWKFCxYuStcYg',
                url: 'https://www.datocms-assets.com/160242/1710227315-facebook-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            }
          ]
        } as Query['footer']
      }
    }

    case 'e-sign-agreement': {
      return {
        legalPage: {
          slug: 'e-sign-agreement',
          external: '',
          body: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'This E-Sign Agreement ("Agreement") outlines the terms and conditions for receiving electronic Communications and using electronic signatures in your relationship with Interledger Foundation and its affiliates and third-party service providers (collectively, "Interledger"). By agreeing to this Agreement, you consent to the electronic delivery of Communications and the use of electronic signatures.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      { type: 'span', marks: ['strong'], value: 'Definitions' }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'bulleted',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  '"Services" refers to the products, software, and services offered by Interledger Foundation.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  '"Communications" include any communications, notices, agreements, billing statements, or disclosures provided by Interledger in relation to our Services.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      {
                        type: 'span',
                        marks: ['strong'],
                        value:
                          'Electronic Delivery of Communications and Use of Electronic Signatures'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Under this Agreement, Interledger may provide all Communications electronically by email, by text message, or by making them accessible via Interledger websites. We may also use electronic signatures and obtain them from you.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      {
                        type: 'span',
                        marks: ['strong'],
                        value: 'System Requirements'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'To access and retain electronic Communications, you will need the following:'
                      }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'bulleted',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'A computer or mobile device with Internet or mobile connectivity.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'A recent web browser that includes 256-bit encryption.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'The browser must have cookies enabled. Use of browser extensions may impair full website functionality.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Access to the email address used to create an account for Services.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Sufficient storage space to save Communications and/or a printer to print them.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'If you use a spam filter that blocks or re-routes emails from senders not listed in your email address book, you must add '
                              },
                              {
                                url: 'mailto:hello@interledger.app',
                                type: 'link',
                                children: [
                                  {
                                    type: 'span',
                                    marks: ['strong'],
                                    value: 'hello@interledger.app'
                                  }
                                ]
                              },
                              {
                                type: 'span',
                                value: ' to your email address book.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      {
                        type: 'span',
                        marks: ['strong'],
                        value: 'User Responsibilities'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      { type: 'span', value: 'It is your responsibility to:' }
                    ]
                  },
                  {
                    type: 'list',
                    style: 'bulleted',
                    children: [
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Ensure that your system meets the requirements outlined above to access and retain electronic Communications.'
                              }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'listItem',
                        children: [
                          {
                            type: 'paragraph',
                            children: [
                              {
                                type: 'span',
                                value:
                                  'Regularly check your email and Interledger account for new Communications.'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      {
                        type: 'span',
                        marks: ['strong'],
                        value: 'Consent Validity'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Your consent is considered valid from the moment you agree to this Agreement until you withdraw your consent or the Agreement is terminated.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      {
                        type: 'span',
                        marks: ['strong'],
                        value:
                          'Withdrawal of Consent to Electronic Communications'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'You may withdraw your consent to receive electronic Communications at any time, by sending an email to '
                      },
                      {
                        url: 'mailto:support@interledger.app',
                        type: 'link',
                        children: [
                          {
                            type: 'span',
                            marks: ['strong'],
                            value: 'support@interldger.app'
                          }
                        ]
                      },
                      {
                        type: 'span',
                        value:
                          '. However, withdrawal of your consent may result in termination of your access to Services. Any withdrawal of your consent will be effective after a reasonable period of time for processing your request, and Interledger will confirm your withdrawal of consent and its effective date in writing.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      {
                        type: 'span',
                        marks: ['strong'],
                        value: 'Termination and Consequences'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Interledger may terminate this Agreement if you withdraw your consent to receive electronic Communications or if you violate the terms of the Agreement. Termination of the Agreement may result in the loss of access to Services and any associated data.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      {
                        type: 'span',
                        marks: ['strong'],
                        value: 'Governing Law and Jurisdiction'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'This Agreement shall be governed by and construed in accordance with the laws of the state of Wyoming, without regard to its conflict of law provisions.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      {
                        type: 'span',
                        marks: ['strong'],
                        value: 'Changes to the Agreement'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'Interledger reserves the right to update or modify this Agreement at any time. We will notify you of any changes to the Agreement via email or through the Services. Your continued use of the Services after receiving notice of any changes constitutes your acceptance of the revised Agreement.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      {
                        type: 'span',
                        marks: ['strong'],
                        value: 'Updating Your Email Address'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'You can change your email address by sending an email to '
                      },
                      {
                        url: 'mailto:support@interledger.app',
                        type: 'link',
                        children: [
                          {
                            type: 'span',
                            marks: ['strong'],
                            value: 'support@interledger.app'
                          }
                        ]
                      },
                      {
                        type: 'span',
                        value: ' or updating it yourself through the Services.'
                      }
                    ]
                  },
                  {
                    type: 'heading',
                    level: 3,
                    children: [
                      {
                        type: 'span',
                        marks: ['strong'],
                        value: 'Contact Information'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'If you have any questions or concerns about this Agreement, please contact us at '
                      },
                      {
                        url: 'mailto:support@interledger.app',
                        type: 'link',
                        children: [
                          {
                            type: 'span',
                            marks: ['strong'],
                            value: 'support@interledger.app'
                          }
                        ]
                      },
                      { type: 'span', marks: ['strong'], value: '.' }
                    ]
                  }
                ]
              }
            },
            __typename: 'LegalPageModelBodyField'
          },
          id: '125342881',
          title: 'E-Sign Agreement',
          _publishedAt: '2024-07-15T16:07:25+02:00',
          _seoMetaTags: [
            {
              tag: 'title',
              attributes: null,
              content: 'E-Sign Agreement',
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:title', content: 'E-Sign Agreement' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                name: 'twitter:title',
                content: 'E-Sign Agreement'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { name: 'twitter:description', content: '.' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:locale', content: 'en' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:type', content: 'article' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'og:site_name', content: 'Interledger' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                property: 'article:modified_time',
                content: '2025-04-06T16:23:03Z'
              },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: { property: 'article:publisher', content: '' },
              content: null,
              __typename: 'Tag'
            },
            {
              tag: 'meta',
              attributes: {
                name: 'twitter:card',
                content: 'summary_large_image'
              },
              content: null,
              __typename: 'Tag'
            }
          ],
          __typename: 'LegalPageRecord'
        } as Query['legalPage'],
        footer: {
          __typename: 'FooterRecord',
          id: '121270040',
          logo: {
            id: 'dffIomSsTfCkd-b3vjChtA',
            url: 'https://www.datocms-assets.com/160242/1721916494-interledger_icon.svg',
            height: 101,
            width: 101,
            __typename: 'FileField'
          },
          column1Title: 'Menu',
          column1: [
            {
              id: '125075096',
              displayText: 'Contact',
              url: 'https://interledger.app/contact',
              __typename: 'LinkRecord'
            }
          ],
          column2Title: 'Ecosystem',
          column2: [
            {
              id: '121270066',
              displayText: 'Interledger Foundation',
              url: 'https://interledger.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075093',
              displayText: 'Web monetization',
              url: 'https://webmonetization.org/',
              __typename: 'LinkRecord'
            },
            {
              id: '125075094',
              displayText: 'Open Payments',
              url: 'https://openpayments.dev/',
              __typename: 'LinkRecord'
            }
          ],
          column3Title: 'Legal',
          column3: [
            {
              id: '121270067',
              displayText: 'Legal Agreements',
              url: 'https://interledger.app/legal',
              __typename: 'LinkRecord'
            }
          ],
          legalText: {
            value: {
              schema: 'dast',
              document: {
                type: 'root',
                children: [
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          '© 2024 Interledger Inc and the Interledger Foundation.'
                      }
                    ]
                  },
                  {
                    type: 'paragraph',
                    children: [
                      {
                        type: 'span',
                        value:
                          'The Interledger name and logo are the property of the'
                      },
                      {
                        url: 'https://interledger.org',
                        type: 'link',
                        children: [
                          { type: 'span', value: ' Interledger Foundation' }
                        ]
                      },
                      {
                        type: 'span',
                        value:
                          '. The Interledger app is powered by Interledger on behalf of the Interledger Foundation as a service to the Interledger community. Interledger Inc is not a bank. Interledger provides a technology platform and all payments and banking services are provided by our partners who are appropriately licensed.'
                      }
                    ]
                  }
                ]
              }
            },
            __typename: 'FooterModelLegalTextField'
          },
          socialIcons: [
            {
              id: '124003009',
              url: 'https://x.com/Interledger',
              icon: {
                id: 'If7vlze6QAeqjixiyS922g',
                url: 'https://www.datocms-assets.com/160242/1710406464-x-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089438',
              url: 'https://www.linkedin.com/company/interledger-foundation/',
              icon: {
                id: '52090585',
                url: 'https://www.datocms-assets.com/160242/1685976901-icon-linkedin.svg',
                height: 16,
                width: 16,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: '126089519',
              url: 'https://www.youtube.com/@InterledgerFoundation',
              icon: {
                id: '52090589',
                url: 'https://www.datocms-assets.com/160242/1685977020-icon-youtube.svg',
                height: 13,
                width: 19,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'PCadAp0bSRW4_0v6BTeiKg',
              url: 'https://www.instagram.com/interledgerfoundation/',
              icon: {
                id: 'dfWXhnzmQT68juUEGsvCkw',
                url: 'https://www.datocms-assets.com/160242/1710227274-instagram-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            },
            {
              id: 'JFxCqHyWQCyJbpfrhMbeJQ',
              url: 'https://www.facebook.com/interledger',
              icon: {
                id: 'YhE1LNyHRSWKFCxYuStcYg',
                url: 'https://www.datocms-assets.com/160242/1710227315-facebook-white-icon.svg',
                height: 24,
                width: 24,
                __typename: 'FileField'
              },
              __typename: 'SocialIconRecord'
            }
          ]
        } as Query['footer']
      }
    }

    default: {
      return {
        legalPage: null,
        footer: null
      }
    }
  }
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
