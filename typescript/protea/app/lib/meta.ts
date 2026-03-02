import type { LoaderFunction, MetaFunction } from 'react-router';

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
