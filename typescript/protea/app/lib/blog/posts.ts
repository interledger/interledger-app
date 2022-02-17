import path, { join } from 'path'
import { readdir, readFile, stat } from 'fs/promises'
import * as authors from './authors'
import { DateTime } from 'luxon'

const postsDirectory = join(process.cwd(), './pages/blog')

export async function getAllPosts() {
  const paths = await walk(postsDirectory)
  const mdxPaths = paths.filter((path) => path.indexOf('blog.mdx') > -1)

  const mdxMeta = await Promise.all(
    mdxPaths.map(async (path) => {
      return await getFileMeta(path)
    })
  )

  return mdxMeta
    .filter((meta) => !meta.preview)
    .sort(
      (post1, post2) =>
        DateTime.fromISO(post2.date).toSeconds() -
        DateTime.fromISO(post1.date).toSeconds()
    )
}

/**
 * This function extracts the metadata object from an mdx file on the fly.
 * @param path
 */
async function getFileMeta(path: string) {
  const fileContents = await readFile(path, 'utf8')
  const metaRegex = /export const meta = \{([\s\S]*?)\}\n\n/g
  let metaString = fileContents.match(metaRegex)
  if (metaString) {
    let newStr = ''
    metaString[0] = metaString[0].slice(20) // remove leading 'export const meta = '
    for (let author in authors) {
      newStr = metaString[0].replace(
        author as string,
        // @ts-ignore
        JSON.stringify(authors[author])
      )
    }
    // test converts string js object (non JSON) to js object
    // eslint-disable-next-line no-new-func
    let test = new Function('return ' + newStr)
    return test()
  }
  return null
}

/**
 * Recursively list all files in a directory.
 * @param dir
 */
const walk = async (dir: string): Promise<string[]> => {
  let fileList: string[] = []

  const files = await readdir(dir)
  for (const file of files) {
    const p = path.join(dir, file)
    if ((await stat(p)).isDirectory()) {
      fileList = [...fileList, ...(await walk(p))]
    } else {
      fileList.push(p)
    }
  }

  return fileList
}
