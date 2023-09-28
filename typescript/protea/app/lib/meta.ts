import type {
  LoaderFunction,
  MetaDescriptor,
  MetaFunction
} from '@remix-run/node'
import type { Location } from '@remix-run/react'
import type { Tag } from '~/generated/dato-cms-graphql'

export const mergeMeta = <
  Loader extends LoaderFunction | unknown = unknown,
  ParentsLoaders extends Record<string, LoaderFunction | unknown> = Record<
    string,
    unknown
  >
>(
  leafMetaFn: MetaFunction<Loader, ParentsLoaders>
): MetaFunction<Loader, ParentsLoaders> => {
  return (arg) => {
    let leafMeta = leafMetaFn(arg)

    return arg.matches.reduceRight((acc, match) => {
      for (let parentMeta of match.meta) {
        let index = acc.findIndex(
          (meta) =>
            ('name' in meta &&
              'name' in parentMeta &&
              meta.name === parentMeta.name) ||
            ('property' in meta &&
              'property' in parentMeta &&
              meta.property === parentMeta.property) ||
            ('title' in meta && 'title' in parentMeta)
        )
        if (index == -1) {
          // Parent meta not found in acc, so add it
          acc.push(parentMeta)
        }
      }
      return acc
    }, leafMeta)
  }
}

export function datoMeta(
  metaTags?: Array<Tag>,
  location?: Location
): MetaDescriptor[] {
  const locationTags = [
    {
      name: 'og:url',
      content: `https://fynbos.app${location?.pathname}`
    },
    {
      name: 'twitter:url',
      content: `https://fynbos.app${location?.pathname}`
    }
  ]

  if (!metaTags) {
    return [...locationTags]
  }

  let tags: MetaDescriptor[] = locationTags
  for (const metaTag of metaTags) {
    if (metaTag.tag === 'title' && metaTag.content) {
      tags.push({ title: metaTag.content })
    } else if (metaTag.tag === 'meta' && metaTag.attributes) {
      tags.push(metaTag.attributes)
    }
  }
  return tags
}
