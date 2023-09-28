import type { MetaDescriptor, MetaFunction } from '@remix-run/node'
import type { Tag } from '~/generated/dato-cms-graphql'

export const mergeMeta = <MergeMetaArgs>(
  overrideFn: MetaFunction<MergeMetaArgs>,
  appendFn?: MetaFunction<MergeMetaArgs>
): MetaFunction<MergeMetaArgs> => {
  return (arg) => {
    // get meta from parent routes
    let mergedMeta = arg.matches.reduce((acc, match: any) => {
      return acc.concat(match.meta || [])
    }, [] as MetaDescriptor[])

    // replace any parent meta with the same name or property with the override
    let overrides = overrideFn(arg)
    for (let override of overrides) {
      let index = mergedMeta.findIndex(
        (meta) =>
          ('name' in meta &&
            'name' in override &&
            meta.name === override.name) ||
          ('property' in meta &&
            'property' in override &&
            meta.property === override.property) ||
          ('title' in meta && 'title' in override)
      )
      if (index !== -1) {
        mergedMeta.splice(index, 1, override)
      }
    }

    // append any additional meta
    if (appendFn) {
      mergedMeta = mergedMeta.concat(appendFn(arg))
    }

    return mergedMeta
  }
}

export function datoMeta(metaTags?: Array<Tag>): MetaDescriptor[] {
  if (!metaTags) {
    return []
  }

  let tags: MetaDescriptor[] = []
  for (const metaTag of metaTags) {
    if (metaTag.tag === 'title' && metaTag.content) {
      tags.push({ title: metaTag.content })
    } else if (metaTag.tag === 'meta' && metaTag.attributes) {
      tags.push(metaTag.attributes)
    }
  }
  return tags
}
